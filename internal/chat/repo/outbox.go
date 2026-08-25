package repo

import (
	"context"
	"nurture/internal/chat/repo/dao"
	"time"
)

func (r *ChatRepo) ListPendingOutbox(ctx context.Context, now int64, staleBefore int64, limit int) ([]ChatOutboxEvent, error) {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if staleBefore <= 0 {
		staleBefore = now
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := r.dao.ClaimPendingChatEventOutbox(ctx, dao.ClaimPendingChatEventOutboxParams{
		ClaimedAt:   now,
		RetryBefore: now,
		StaleBefore: staleBefore,
		ClaimLimit:  int32(limit),
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]ChatOutboxEvent, 0, len(rows))
	for _, v := range rows {
		items = append(items, ChatOutboxEvent{
			ID:         v.ID,
			EventID:    v.EventID,
			RoutingKey: v.RoutingKey,
			Payload:    v.Payload,
			Attempts:   v.Attempts,
			Ctime:      v.Ctime,
		})
	}
	return items, nil
}

func (r *ChatRepo) MarkOutboxPublished(ctx context.Context, id int64, now int64) error {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	_, err := r.dao.MarkChatEventOutboxPublished(ctx, dao.MarkChatEventOutboxPublishedParams{
		ID:          id,
		PublishedAt: now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *ChatRepo) MarkOutboxFailed(ctx context.Context, id int64, nextRetryAt int64, maxAttempts int32, now int64) error {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	_, err := r.dao.MarkChatEventOutboxFailed(ctx, dao.MarkChatEventOutboxFailedParams{
		ID:          id,
		NextRetryAt: nextRetryAt,
		Attempts:    maxAttempts,
		Utime:       now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}
