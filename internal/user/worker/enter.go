package worker

import (
	"context"
	"errors"
	"nurture/internal/pkg/rabbitmqx"
	"nurture/internal/pkg/retryx"
	"nurture/internal/pkg/zapx"
	"nurture/internal/user/constant"
	"nurture/internal/user/event"
	"nurture/internal/user/repo"
	"time"

	"go.uber.org/zap"
)

type EventPublisher interface {
	DeclareTopicExchange(name string) error
	Publish(ctx context.Context, exchange, routingKey string, msg rabbitmqx.PublishMessage) error
}

type OutboxStore interface {
	ListPendingOutbox(ctx context.Context, now int64, staleBefore int64, limit int) ([]repo.UserOutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, id int64, now int64) error
	MarkOutboxFailed(ctx context.Context, id int64, nextRetryAt int64, maxAttempts int32, now int64) error
}

var ErrNilOutboxEvent = errors.New("nil user outbox event")

type Worker struct {
	repo OutboxStore
	bus  EventPublisher
	log  *zap.SugaredLogger
}

func NewOutboxWorker(repo OutboxStore, bus EventPublisher, log *zap.SugaredLogger) *Worker {
	return &Worker{
		repo: repo,
		bus:  bus,
		log:  zapx.OrNop(log),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w == nil || w.repo == nil || w.bus == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(constant.OutboxPollInterval)
		defer ticker.Stop()
		for {
			w.publishBatch(ctx)
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}

func (w *Worker) publishBatch(ctx context.Context) {
	if err := w.bus.DeclareTopicExchange(event.Exchange); err != nil {
		w.log.Error(err)
		return
	}
	now := time.Now().UnixMilli()
	staleBefore := time.Now().Add(-constant.OutboxClaimTimeout).UnixMilli()
	items, err := w.repo.ListPendingOutbox(ctx, now, staleBefore, constant.OutboxBatchSize)
	if err != nil {
		w.log.Error(err)
		return
	}
	for _, item := range items {
		if ctx.Err() != nil {
			return
		}
		if err := w.Handle(ctx, item); err != nil {
			w.log.Error(err)
		}
	}
}

func (w *Worker) Handle(ctx context.Context, item repo.UserOutboxEvent) error {
	if w == nil || w.repo == nil || w.bus == nil || item.EventID == "" || item.Payload == "" {
		return ErrNilOutboxEvent
	}
	now := time.Now().UnixMilli()
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := w.bus.Publish(publishCtx, event.Exchange, item.RoutingKey, rabbitmqx.PublishMessage{
		MessageID:   item.EventID,
		ContentType: event.ContentTypeJSON,
		Body:        []byte(item.Payload),
	})
	if err != nil {
		nextRetryAt := time.Now().Add(retryx.ExponentialBackoff(
			constant.OutboxRetryBaseDelay,
			int64(item.Attempts+1),
			constant.OutboxRetryMaxDelay,
		)).UnixMilli()
		if markErr := w.repo.MarkOutboxFailed(ctx, item.ID, nextRetryAt, constant.OutboxMaxAttempts, now); markErr != nil {
			return markErr
		}
		return err
	}
	return w.repo.MarkOutboxPublished(ctx, item.ID, now)
}
