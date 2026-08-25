package rabbitmqx

import (
	"context"
	"errors"
	"fmt"
	"nurture/internal/config"
	"nurture/internal/pkg/zapx"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type Client struct {
	conn *amqp.Connection
	log  *zap.SugaredLogger
}

type PublishMessage struct {
	MessageID   string
	ContentType string
	Body        []byte
}

type RetryConfig struct {
	Delay       time.Duration
	MaxAttempts int64
}

type ConsumeConfig struct {
	Exchange    string
	Queue       string
	Consumer    string
	RoutingKeys []string
	Retry       RetryConfig
}

type Delivery struct {
	MessageID   string
	RoutingKey  string
	Redelivered bool
	Attempt     int64
	Body        []byte
}

func InitRabbitMQ(log *zap.SugaredLogger) *Client {
	if !config.Conf.RabbitMQ.Enable {
		return nil
	}
	client, err := NewClient(config.Conf.RabbitMQ.DSN(), log)
	if err != nil {
		panic(fmt.Sprintf("rabbitmq init error: %v", err))
	}
	return client
}

func NewClient(dsn string, log *zap.SugaredLogger) (*Client, error) {
	conn, err := amqp.Dial(dsn)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, log: zapx.OrNop(log)}, nil
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) DeclareTopicExchange(name string) error {
	if c == nil || c.conn == nil {
		return errors.New("rabbitmq client is nil")
	}
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	return ch.ExchangeDeclare(
		name,
		amqp.ExchangeTopic,
		true,
		false,
		false,
		false,
		nil,
	)
}

func (c *Client) Publish(ctx context.Context, exchange, routingKey string, msg PublishMessage) error {
	if c == nil || c.conn == nil {
		return errors.New("rabbitmq client is nil")
	}
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.Confirm(false); err != nil {
		return err
	}
	confirms := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  msg.ContentType,
		DeliveryMode: amqp.Persistent,
		MessageId:    msg.MessageID,
		Timestamp:    time.Now(),
		Body:         msg.Body,
	}); err != nil {
		return err
	}
	select {
	case confirm := <-confirms:
		if !confirm.Ack {
			return errors.New("rabbitmq publish not acknowledged")
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (c *Client) Consume(ctx context.Context, cfg ConsumeConfig, handle func(context.Context, Delivery) error) error {
	if c == nil || c.conn == nil {
		return errors.New("rabbitmq client is nil")
	}
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()
	if err := ch.ExchangeDeclare(cfg.Exchange, amqp.ExchangeTopic, true, false, false, false, nil); err != nil {
		return err
	}
	topology := newConsumeTopology(cfg)
	if topology.RetryEnabled {
		if err := declareRetryTopology(ch, cfg, topology); err != nil {
			return err
		}
	}
	q, err := ch.QueueDeclare(cfg.Queue, false, true, true, false, topology.MainQueueArgs)
	if err != nil {
		return err
	}
	for _, routingKey := range cfg.RoutingKeys {
		if err := ch.QueueBind(q.Name, routingKey, cfg.Exchange, false, nil); err != nil {
			return err
		}
	}
	deliveries, err := ch.Consume(q.Name, cfg.Consumer, false, true, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return nil
			}
			attempt := deliveryAttempt(msg.Headers, q.Name)
			err := handle(ctx, Delivery{
				MessageID:   msg.MessageId,
				RoutingKey:  msg.RoutingKey,
				Redelivered: msg.Redelivered,
				Attempt:     attempt,
				Body:        msg.Body,
			})
			if err != nil {
				c.log.Error(err)
				if isDiscard(err) {
					_ = msg.Ack(false)
					continue
				}
				if topology.RetryEnabled {
					if attempt >= topology.MaxAttempts {
						if err := publishDeadLetter(ctx, ch, topology, msg, err); err != nil {
							c.log.Error(err)
							_ = msg.Nack(false, false)
							continue
						}
						_ = msg.Ack(false)
						continue
					}
					_ = msg.Nack(false, false)
					continue
				}
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}
