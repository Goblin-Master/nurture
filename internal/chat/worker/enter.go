package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/event"
	"nurture/internal/chat/session"
	"nurture/internal/pkg/rabbitmqx"
	"os"
	"time"

	"go.uber.org/zap"
)

type Consumer interface {
	Consume(ctx context.Context, cfg rabbitmqx.ConsumeConfig, handle func(context.Context, rabbitmqx.Delivery) error) error
}

type Worker struct {
	consumer Consumer
	hub      *session.Hub
	log      *zap.SugaredLogger
}

type directOutMessage struct {
	Op      string            `json:"op"`
	Message directMessageBody `json:"message"`
}

type directMessageBody struct {
	MessageID  string `json:"message_id"`
	FromUserID string `json:"from_user_id"`
	ToUserID   string `json:"to_user_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Ctime      int64  `json:"ctime"`
}

func NewWorker(consumer Consumer, hub *session.Hub, log *zap.SugaredLogger) *Worker {
	return &Worker{
		consumer: consumer,
		hub:      hub,
		log:      log,
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w == nil || w.consumer == nil || w.hub == nil {
		return
	}
	go func() {
		for {
			err := w.consumer.Consume(ctx, rabbitmqx.ConsumeConfig{
				Exchange: event.Exchange,
				Queue:    instanceQueueName(),
				Consumer: event.ConsumerName,
				RoutingKeys: []string{
					event.RoutingKeyDirect,
					event.RoutingKeyGroup,
				},
				Retry: rabbitmqx.RetryConfig{
					Delay:       constant.ConsumerRetryDelay,
					MaxAttempts: constant.ConsumerMaxAttempts,
				},
			}, w.handle)
			if ctx.Err() != nil {
				return
			}
			if w.log != nil && err != nil {
				w.log.Error(err)
			}
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Second):
			}
		}
	}()
}

func (w *Worker) handle(_ context.Context, delivery rabbitmqx.Delivery) error {
	switch delivery.RoutingKey {
	case event.RoutingKeyDirect:
		return w.handleDirect(delivery)
	case event.RoutingKeyGroup:
		return w.handleGroup(delivery)
	default:
		return rabbitmqx.Discard(fmt.Errorf("unknown chat event routing key: %s", delivery.RoutingKey))
	}
}

func (w *Worker) handleDirect(delivery rabbitmqx.Delivery) error {
	var msg event.DirectMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return rabbitmqx.Discard(fmt.Errorf("decode direct chat event: %w", err))
	}
	eventID := msg.EventID
	if eventID == "" {
		eventID = delivery.MessageID
	}
	payload := msg.Payload
	if payload == "" {
		out, err := json.Marshal(directOutMessage{
			Op: "new_message",
			Message: directMessageBody{
				MessageID:  msg.MessageID,
				FromUserID: msg.FromUserID,
				ToUserID:   msg.ToUserID,
				Type:       msg.Type,
				Content:    msg.Content,
				Ctime:      msg.Ctime,
			},
		})
		if err != nil {
			return rabbitmqx.Discard(fmt.Errorf("build direct chat payload: %w", err))
		}
		payload = string(out)
	}
	w.hub.DeliverToUser(constant.ChannelDirect, msg.ToUserID, eventID, []byte(payload))
	return nil
}

func (w *Worker) handleGroup(delivery rabbitmqx.Delivery) error {
	var msg event.GroupMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return rabbitmqx.Discard(fmt.Errorf("decode group chat event: %w", err))
	}
	eventID := msg.EventID
	if eventID == "" {
		eventID = delivery.MessageID
	}
	if msg.Payload == "" {
		return rabbitmqx.Discard(fmt.Errorf("missing group chat payload: %s", eventID))
	}
	w.hub.DeliverToRoom(msg.GroupID, eventID, []byte(msg.Payload))
	return nil
}

func instanceQueueName() string {
	host, err := os.Hostname()
	if err != nil || host == "" {
		host = "unknown"
	}
	return fmt.Sprintf("%s.%s.%d", event.QueueNamePrefix, host, os.Getpid())
}
