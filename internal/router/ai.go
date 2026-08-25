package router

import (
	"context"
	"errors"
	ailogic "nurture/internal/ai/logic"
	babyrepo "nurture/internal/baby/repo"
)

type aiGrowthReader struct {
	babyRepo babyrepo.IBabyRepo
}

func newAIGrowthReader(babyRepo babyrepo.IBabyRepo) *aiGrowthReader {
	if babyRepo == nil {
		return nil
	}
	return &aiGrowthReader{babyRepo: babyRepo}
}

func (r *aiGrowthReader) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (ailogic.BabyProfile, error) {
	b, err := r.babyRepo.GetBabyByIDAndUser(ctx, babyID, userID)
	if err != nil {
		if errors.Is(err, babyrepo.ErrBabyNotExist) {
			return ailogic.BabyProfile{}, ailogic.ErrBabyNotExist
		}
		return ailogic.BabyProfile{}, err
	}
	return ailogic.BabyProfile{
		BabyID:   b.BabyID.String(),
		Name:     b.Name,
		Gender:   b.Gender,
		Birthday: b.Birthday,
		Avatar:   b.Avatar,
	}, nil
}

func (r *aiGrowthReader) ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, from, to int64) ([]ailogic.GrowthRecord, error) {
	rows, err := r.babyRepo.ListGrowthRecordsByBabyIDBetween(ctx, babyID, from, to)
	if err != nil {
		return nil, err
	}
	ret := make([]ailogic.GrowthRecord, 0, len(rows))
	for _, row := range rows {
		item := ailogic.GrowthRecord{
			RecordTime: row.RecordTime,
		}
		if row.Height.Valid {
			item.Height = ptrFloat64(row.Height.Float64)
		}
		if row.Weight.Valid {
			item.Weight = ptrFloat64(row.Weight.Float64)
		}
		if row.HeadCircumference.Valid {
			item.HeadCircumference = ptrFloat64(row.HeadCircumference.Float64)
		}
		if row.Remark.Valid {
			item.Remark = row.Remark.String
		}
		ret = append(ret, item)
	}
	return ret, nil
}

func ptrFloat64(v float64) *float64 {
	return &v
}
