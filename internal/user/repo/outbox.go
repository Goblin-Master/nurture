package repo

import (
	"context"
	"nurture/internal/user/event"
	"nurture/internal/user/repo/dao"
	"time"
)

type UserOutboxEvent struct {
	ID         int64
	EventID    string
	RoutingKey string
	Payload    string
	Attempts   int32
	Ctime      int64
}

func (ur *UserRepo) CreateUserEventOutbox(ctx context.Context, q *dao.Queries, outbox UserOutboxEvent, now int64) error {
	if outbox.EventID == "" || outbox.RoutingKey == "" || outbox.Payload == "" {
		return ErrParamsType
	}
	if outbox.Ctime <= 0 {
		outbox.Ctime = now
	}
	aff, err := q.CreateUserEventOutbox(ctx, dao.CreateUserEventOutboxParams{
		EventID:    outbox.EventID,
		RoutingKey: outbox.RoutingKey,
		Payload:    outbox.Payload,
		Ctime:      outbox.Ctime,
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	if aff == 0 {
		return ErrDefault
	}
	return nil
}

func (ur *UserRepo) ListPendingOutbox(ctx context.Context, now int64, staleBefore int64, limit int) ([]UserOutboxEvent, error) {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if staleBefore <= 0 {
		staleBefore = now
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := ur.userDao.ClaimPendingUserEventOutbox(ctx, dao.ClaimPendingUserEventOutboxParams{
		ClaimedAt:   now,
		RetryBefore: now,
		StaleBefore: staleBefore,
		ClaimLimit:  int32(limit),
	})
	if err != nil {
		ur.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]UserOutboxEvent, 0, len(rows))
	for _, v := range rows {
		items = append(items, UserOutboxEvent{
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

func (ur *UserRepo) MarkOutboxPublished(ctx context.Context, id int64, now int64) error {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	_, err := ur.userDao.MarkUserEventOutboxPublished(ctx, dao.MarkUserEventOutboxPublishedParams{
		ID:          id,
		PublishedAt: now,
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (ur *UserRepo) MarkOutboxFailed(ctx context.Context, id int64, nextRetryAt int64, maxAttempts int32, now int64) error {
	if now <= 0 {
		now = time.Now().UnixMilli()
	}
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	_, err := ur.userDao.MarkUserEventOutboxFailed(ctx, dao.MarkUserEventOutboxFailedParams{
		ID:          id,
		NextRetryAt: nextRetryAt,
		Attempts:    maxAttempts,
		Utime:       now,
	})
	if err != nil {
		ur.log.Error(err)
		return ErrDefault
	}
	return nil
}

func newPartnerBoundOutbox(fatherUserID, motherUserID string, occurredAt int64) (UserOutboxEvent, error) {
	msg, err := event.NewPartnerBoundMessage(fatherUserID, motherUserID, occurredAt)
	if err != nil {
		return UserOutboxEvent{}, err
	}
	return UserOutboxEvent{
		EventID:    msg.EventID,
		RoutingKey: event.RoutingKeyPartnerBound,
		Payload:    msg.Payload,
		Ctime:      occurredAt,
	}, nil
}
