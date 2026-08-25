package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"nurture/internal/baby/constant"
	"nurture/internal/pkg/rabbitmqx"
	"nurture/internal/pkg/zapx"
	"time"

	"go.uber.org/zap"
)

type Consumer interface {
	Consume(ctx context.Context, cfg rabbitmqx.ConsumeConfig, handle func(context.Context, rabbitmqx.Delivery) error) error
}

type BabyEventLogic interface {
	HandlePartnerBound(ctx context.Context, eventID, fatherUserID, motherUserID string) error
}

type Worker struct {
	consumer Consumer
	logic    BabyEventLogic
	log      *zap.SugaredLogger
}

type partnerBoundMessage struct {
	EventID      string `json:"event_id"`
	FatherUserID string `json:"father_user_id"`
	MotherUserID string `json:"mother_user_id"`
	OccurredAt   int64  `json:"occurred_at"`
}

func NewWorker(consumer Consumer, logic BabyEventLogic, log *zap.SugaredLogger) *Worker {
	return &Worker{
		consumer: consumer,
		logic:    logic,
		log:      zapx.OrNop(log),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w == nil || w.consumer == nil || w.logic == nil {
		return
	}
	go func() {
		for {
			err := w.consumer.Consume(ctx, rabbitmqx.ConsumeConfig{
				Exchange:     constant.UserEventExchange,
				Queue:        constant.PartnerBoundQueue,
				Consumer:     constant.PartnerBoundConsumer,
				RoutingKeys:  []string{constant.UserPartnerBoundRoutingKey},
				DurableQueue: true,
				Retry: rabbitmqx.RetryConfig{
					Delay:       constant.ConsumerRetryDelay,
					MaxAttempts: constant.ConsumerMaxAttempts,
				},
			}, w.Handle)
			if ctx.Err() != nil {
				return
			}
			if err != nil {
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

func (w *Worker) Handle(ctx context.Context, delivery rabbitmqx.Delivery) error {
	if delivery.RoutingKey != constant.UserPartnerBoundRoutingKey {
		return rabbitmqx.Discard(fmt.Errorf("unknown user event routing key: %s", delivery.RoutingKey))
	}
	var msg partnerBoundMessage
	if err := json.Unmarshal(delivery.Body, &msg); err != nil {
		return rabbitmqx.Discard(fmt.Errorf("decode partner bound event: %w", err))
	}
	if msg.EventID == "" {
		msg.EventID = delivery.MessageID
	}
	if msg.EventID == "" || msg.FatherUserID == "" || msg.MotherUserID == "" {
		return rabbitmqx.Discard(fmt.Errorf("invalid partner bound event: %s", msg.EventID))
	}
	if w == nil || w.logic == nil {
		return fmt.Errorf("baby event logic is nil")
	}
	return w.logic.HandlePartnerBound(ctx, msg.EventID, msg.FatherUserID, msg.MotherUserID)
}
