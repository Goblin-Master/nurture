package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/repo"
	"nurture/internal/pkg/zapx"

	"go.uber.org/zap"
)

type PartnerBoundEventStore interface {
	HandlePartnerBoundEvent(ctx context.Context, eventID, fatherUserID, motherUserID string) (bool, error)
}

type BabyEventLogic struct {
	store PartnerBoundEventStore
	log   *zap.SugaredLogger
}

func NewBabyEventLogic(store PartnerBoundEventStore, log *zap.SugaredLogger) *BabyEventLogic {
	return &BabyEventLogic{
		store: store,
		log:   zapx.OrNop(log),
	}
}

func (l *BabyEventLogic) HandlePartnerBound(ctx context.Context, eventID, fatherUserID, motherUserID string) error {
	if eventID == "" || fatherUserID == "" || motherUserID == "" {
		return ErrParamsType
	}
	_, err := l.store.HandlePartnerBoundEvent(ctx, eventID, fatherUserID, motherUserID)
	if err != nil {
		if errors.Is(err, repo.ErrParamsType) {
			return ErrParamsType
		}
		if errors.Is(err, repo.ErrDefault) {
			return ErrDefault
		}
		l.log.Error(err)
		return ErrDefault
	}
	return nil
}
