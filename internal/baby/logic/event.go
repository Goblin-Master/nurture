package logic

import (
	"context"
	"errors"
	"nurture/internal/baby/repo"
	"nurture/internal/pkg/zapx"

	"go.uber.org/zap"
)

type BabyEventRepo interface {
	HandlePartnerBoundEvent(ctx context.Context, eventID, fatherUserID, motherUserID string) (bool, error)
}

type IBabyEventLogic interface {
	HandlePartnerBound(ctx context.Context, eventID, fatherUserID, motherUserID string) error
}

type BabyEventLogic struct {
	repo BabyEventRepo
	log  *zap.SugaredLogger
}

func NewBabyEventLogic(repo BabyEventRepo, log *zap.SugaredLogger) *BabyEventLogic {
	return &BabyEventLogic{
		repo: repo,
		log:  zapx.OrNop(log),
	}
}

func (l *BabyEventLogic) HandlePartnerBound(ctx context.Context, eventID, fatherUserID, motherUserID string) error {
	if eventID == "" || fatherUserID == "" || motherUserID == "" {
		return ErrParamsType
	}
	_, err := l.repo.HandlePartnerBoundEvent(ctx, eventID, fatherUserID, motherUserID)
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

var _ IBabyEventLogic = (*BabyEventLogic)(nil)
