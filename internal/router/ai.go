package router

import (
	"context"
	"errors"
	ailogic "nurture/internal/ai/logic"
	baby "nurture/internal/baby"
)

type babyGrowthClient interface {
	GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (baby.BabyProfile, error)
	ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, from, to int64) ([]baby.GrowthRecord, error)
}

type aiGrowthReader struct {
	babyClient babyGrowthClient
}

func newAIGrowthReader(babyClient babyGrowthClient) *aiGrowthReader {
	if babyClient == nil {
		return nil
	}
	return &aiGrowthReader{babyClient: babyClient}
}

func (r *aiGrowthReader) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (ailogic.BabyProfile, error) {
	b, err := r.babyClient.GetBabyByIDAndUser(ctx, babyID, userID)
	if err != nil {
		if errors.Is(err, baby.ErrBabyNotExist) {
			return ailogic.BabyProfile{}, ailogic.ErrBabyNotExist
		}
		return ailogic.BabyProfile{}, err
	}
	return ailogic.BabyProfile{
		BabyID:   b.BabyID,
		Name:     b.Name,
		Gender:   b.Gender,
		Birthday: b.Birthday,
		Avatar:   b.Avatar,
	}, nil
}

func (r *aiGrowthReader) ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, from, to int64) ([]ailogic.GrowthRecord, error) {
	rows, err := r.babyClient.ListGrowthRecordsByBabyIDBetween(ctx, babyID, from, to)
	if err != nil {
		return nil, err
	}
	ret := make([]ailogic.GrowthRecord, 0, len(rows))
	for _, row := range rows {
		ret = append(ret, ailogic.GrowthRecord{
			RecordTime:        row.RecordTime,
			Height:            row.Height,
			Weight:            row.Weight,
			HeadCircumference: row.HeadCircumference,
			Remark:            row.Remark,
		})
	}
	return ret, nil
}
