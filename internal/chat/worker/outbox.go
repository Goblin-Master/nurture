package worker

import (
	"context"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/event"
	"nurture/internal/chat/repo"
	"nurture/internal/pkg/rabbitmqx"
	"nurture/internal/pkg/zapx"
	"time"

	"go.uber.org/zap"
)

type OutboxRepo interface {
	ListPendingOutbox(ctx context.Context, now int64, staleBefore int64, limit int) ([]repo.ChatOutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, id int64, now int64) error
	MarkOutboxFailed(ctx context.Context, id int64, nextRetryAt int64, maxAttempts int32, now int64) error
}

type OutboxWorker struct {
	repo OutboxRepo
	bus  event.Bus
	log  *zap.SugaredLogger
}

func NewOutboxWorker(repo OutboxRepo, bus event.Bus, log *zap.SugaredLogger) *OutboxWorker {
	return &OutboxWorker{
		repo: repo,
		bus:  bus,
		log:  zapx.OrNop(log),
	}
}

func (w *OutboxWorker) Start(ctx context.Context) {
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

func (w *OutboxWorker) publishBatch(ctx context.Context) {
	if err := w.bus.DeclareTopicExchange(event.Exchange); err != nil {
		w.logError(err)
		return
	}
	now := time.Now().UnixMilli()
	staleBefore := time.Now().Add(-constant.OutboxClaimTimeout).UnixMilli()
	items, err := w.repo.ListPendingOutbox(ctx, now, staleBefore, constant.OutboxBatchSize)
	if err != nil {
		w.logError(err)
		return
	}
	for _, item := range items {
		if ctx.Err() != nil {
			return
		}
		w.publishOne(ctx, item)
	}
}

func (w *OutboxWorker) publishOne(ctx context.Context, item repo.ChatOutboxEvent) {
	now := time.Now().UnixMilli()
	publishCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := w.bus.Publish(publishCtx, event.Exchange, item.RoutingKey, rabbitmqx.PublishMessage{
		MessageID:   item.EventID,
		ContentType: event.ContentTypeJSON,
		Body:        []byte(item.Payload),
	})
	if err != nil {
		w.logError(err)
		nextRetryAt := time.Now().Add(outboxRetryDelay(item.Attempts + 1)).UnixMilli()
		if err := w.repo.MarkOutboxFailed(ctx, item.ID, nextRetryAt, constant.OutboxMaxAttempts, now); err != nil {
			w.logError(err)
		}
		return
	}
	if err := w.repo.MarkOutboxPublished(ctx, item.ID, now); err != nil {
		w.logError(err)
	}
}

func outboxRetryDelay(attempt int32) time.Duration {
	if attempt <= 0 {
		attempt = 1
	}
	shift := attempt - 1
	if shift > 5 {
		shift = 5
	}
	return constant.OutboxRetryBaseDelay * time.Duration(1<<shift)
}

func (w *OutboxWorker) logError(err error) {
	if err != nil {
		w.log.Error(err)
	}
}
