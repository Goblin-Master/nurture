package rabbitmqx

import (
	"context"
	"errors"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type consumeTopology struct {
	RetryEnabled   bool
	MaxAttempts    int64
	RetryExchange  string
	RetryQueue     string
	RetryQueueArgs amqp.Table
	DeadExchange   string
	DeadQueue      string
	MainQueueArgs  amqp.Table
}

type discardError struct {
	err error
}

func Discard(err error) error {
	if err == nil {
		return nil
	}
	return discardError{err: err}
}

func (e discardError) Error() string {
	return e.err.Error()
}

func (e discardError) Unwrap() error {
	return e.err
}

func isDiscard(err error) bool {
	var target discardError
	return errors.As(err, &target)
}

func newConsumeTopology(cfg ConsumeConfig) consumeTopology {
	if cfg.Queue == "" || cfg.Retry.Delay <= 0 || cfg.Retry.MaxAttempts <= 1 {
		return consumeTopology{}
	}
	retryName := cfg.Queue + ".retry"
	deadName := cfg.Queue + ".dead"
	return consumeTopology{
		RetryEnabled:  true,
		MaxAttempts:   cfg.Retry.MaxAttempts,
		RetryExchange: retryName,
		RetryQueue:    retryName,
		RetryQueueArgs: amqp.Table{
			"x-message-ttl":             retryDelayMillis(cfg.Retry.Delay),
			"x-dead-letter-exchange":    amqp.DefaultExchange,
			"x-dead-letter-routing-key": cfg.Queue,
		},
		DeadExchange: deadName,
		DeadQueue:    deadName,
		MainQueueArgs: amqp.Table{
			"x-dead-letter-exchange": retryName,
		},
	}
}

func declareRetryTopology(ch *amqp.Channel, cfg ConsumeConfig, topology consumeTopology) error {
	if err := ch.ExchangeDeclare(topology.RetryExchange, amqp.ExchangeTopic, false, true, false, false, nil); err != nil {
		return err
	}
	if err := ch.ExchangeDeclare(topology.DeadExchange, amqp.ExchangeTopic, false, true, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(topology.RetryQueue, false, true, true, false, topology.RetryQueueArgs); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(topology.DeadQueue, false, true, true, false, nil); err != nil {
		return err
	}
	for _, routingKey := range cfg.RoutingKeys {
		if err := ch.QueueBind(topology.RetryQueue, routingKey, topology.RetryExchange, false, nil); err != nil {
			return err
		}
		if err := ch.QueueBind(topology.DeadQueue, routingKey, topology.DeadExchange, false, nil); err != nil {
			return err
		}
	}
	return nil
}

func publishDeadLetter(ctx context.Context, ch *amqp.Channel, topology consumeTopology, msg amqp.Delivery, cause error) error {
	headers := cloneHeaders(msg.Headers)
	if cause != nil {
		headers["x-error"] = cause.Error()
	}
	return ch.PublishWithContext(ctx, topology.DeadExchange, msg.RoutingKey, false, false, amqp.Publishing{
		Headers:         headers,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationId:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageId:       msg.MessageId,
		Timestamp:       time.Now(),
		Type:            msg.Type,
		UserId:          msg.UserId,
		AppId:           msg.AppId,
		Body:            msg.Body,
	})
}

func deliveryAttempt(headers amqp.Table, queue string) int64 {
	return deathCount(headers, queue) + 1
}

func deathCount(headers amqp.Table, queue string) int64 {
	v, ok := headers["x-death"]
	if !ok {
		return 0
	}
	var maxCount int64
	for _, item := range deathItems(v) {
		table, ok := item.(amqp.Table)
		if !ok {
			continue
		}
		if table["queue"] != queue {
			continue
		}
		if count, ok := toInt64(table["count"]); ok && count > maxCount {
			maxCount = count
		}
	}
	return maxCount
}

func deathItems(v any) []any {
	switch items := v.(type) {
	case []any:
		return items
	case []amqp.Table:
		out := make([]any, 0, len(items))
		for _, item := range items {
			out = append(out, item)
		}
		return out
	default:
		return nil
	}
}

func toInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int64:
		return n, true
	case int32:
		return int64(n), true
	case int:
		return int64(n), true
	case int16:
		return int64(n), true
	case int8:
		return int64(n), true
	case uint64:
		if n > uint64(maxInt64) {
			return 0, false
		}
		return int64(n), true
	case uint32:
		return int64(n), true
	case uint:
		if uint64(n) > uint64(maxInt64) {
			return 0, false
		}
		return int64(n), true
	default:
		return 0, false
	}
}

func retryDelayMillis(delay time.Duration) int32 {
	ms := delay.Milliseconds()
	if ms <= 0 {
		return 0
	}
	if ms > maxInt32 {
		return int32(maxInt32)
	}
	return int32(ms)
}

func cloneHeaders(headers amqp.Table) amqp.Table {
	out := make(amqp.Table, len(headers)+1)
	for k, v := range headers {
		out[k] = v
	}
	return out
}

const (
	maxInt32 = int64(1<<31 - 1)
	maxInt64 = uint64(1<<63 - 1)
)
