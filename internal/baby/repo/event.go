package repo

import (
	"context"
	babyconstant "nurture/internal/baby/constant"
	"nurture/internal/baby/repo/dao"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func (r *BabyRepo) HandlePartnerBoundEvent(ctx context.Context, eventID, fatherUserID, motherUserID string) (bool, error) {
	if eventID == "" || fatherUserID == "" || motherUserID == "" {
		return false, ErrParamsType
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	defer tx.Rollback(ctx)

	qtx := r.babyDao.WithTx(tx)
	now := time.Now().UnixMilli()
	created, err := qtx.CreateBabyEventInbox(ctx, dao.CreateBabyEventInboxParams{
		EventID:   eventID,
		EventType: babyconstant.UserPartnerBoundRoutingKey,
		Ctime:     now,
	})
	if err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	if created == 0 {
		return false, nil
	}

	var fatherUUID, motherUUID pgtype.UUID
	if err := fatherUUID.Scan(fatherUserID); err != nil {
		return false, ErrParamsType
	}
	if err := motherUUID.Scan(motherUserID); err != nil {
		return false, ErrParamsType
	}
	if err := r.copyBabies(ctx, qtx, fatherUUID, motherUUID); err != nil {
		return false, err
	}
	if err := r.copyBabies(ctx, qtx, motherUUID, fatherUUID); err != nil {
		return false, err
	}
	if _, err := qtx.MarkBabyEventInboxProcessed(ctx, dao.MarkBabyEventInboxProcessedParams{
		EventID: eventID,
		Utime:   time.Now().UnixMilli(),
	}); err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	if err := tx.Commit(ctx); err != nil {
		r.log.Error(err)
		return false, ErrDefault
	}
	return true, nil
}
