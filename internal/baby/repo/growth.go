package repo

import (
	"context"
	babyconstant "nurture/internal/baby/constant"
	"nurture/internal/baby/repo/cache"
	"nurture/internal/baby/repo/dao"

	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (r *BabyRepo) GetLatestGrowthByBabyID(ctx context.Context, babyID string) (dao.BabyGrowthRecord, error) {
	key := cache.LatestGrowthKey(babyID)
	{
		var cached dao.BabyGrowthRecord
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached, nil
		}
	}
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return dao.BabyGrowthRecord{}, err
	}
	gr, err := r.babyDao.GetLatestGrowthByBabyID(ctx, bid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dao.BabyGrowthRecord{}, ErrBabyGrowthNotExist
		}
		r.log.Error(err)
		return dao.BabyGrowthRecord{}, ErrDefault
	}
	_ = cache.SetJSON(ctx, r.rdb, key, gr, time.Duration(babyconstant.LatestGrowthTTL)*time.Second)
	return gr, nil
}

func (r *BabyRepo) GetGrowthByBabyIDBetween(ctx context.Context, babyID string, start, end int64) (dao.BabyGrowthRecord, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return dao.BabyGrowthRecord{}, err
	}
	gr, err := r.babyDao.GetGrowthByBabyIDBetween(ctx, dao.GetGrowthByBabyIDBetweenParams{
		BabyID:       bid,
		RecordTime:   start,
		RecordTime_2: end,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dao.BabyGrowthRecord{}, ErrBabyGrowthNotExist
		}
		r.log.Error(err)
		return dao.BabyGrowthRecord{}, ErrDefault
	}
	return gr, nil
}

func (r *BabyRepo) ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, start, end int64) ([]dao.BabyGrowthRecord, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListGrowthRecordsByBabyIDBetween(ctx, dao.ListGrowthRecordsByBabyIDBetweenParams{
		BabyID:       bid,
		RecordTime:   start,
		RecordTime_2: end,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	return rows, nil
}

func (r *BabyRepo) UpdateGrowthByRecordID(ctx context.Context, recordID string, recordTime int64, height, weight, headCircumference float64, remark string, updatedBy string) error {
	var rid pgtype.UUID
	if err := rid.Scan(recordID); err != nil {
		return err
	}
	var uid pgtype.UUID
	if err := uid.Scan(updatedBy); err != nil {
		return err
	}
	err := r.babyDao.UpdateGrowthByRecordID(ctx, dao.UpdateGrowthByRecordIDParams{
		RecordID:          rid,
		RecordTime:        recordTime,
		Height:            pgtype.Float8{Float64: height, Valid: height != 0},
		Weight:            pgtype.Float8{Float64: weight, Valid: weight != 0},
		HeadCircumference: pgtype.Float8{Float64: headCircumference, Valid: headCircumference != 0},
		Column6:           remark,
		UpdatedBy:         uid,
		Utime:             recordTime,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	if babyID, err := r.babyDao.GetBabyIDByGrowthRecordID(ctx, rid); err == nil {
		_ = cache.Del(ctx, r.rdb, cache.LatestGrowthKey(babyID))
	}
	return nil
}

func (r *BabyRepo) CreateGrowthRecord(ctx context.Context, babyID, userID string, recordTime int64, height, weight, headCircumference float64, remark string) (string, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", err
	}
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	var rid pgtype.UUID
	if err := rid.Scan(uuid.NewString()); err != nil {
		return "", err
	}
	now := recordTime
	if now <= 0 {
		now = recordTime
	}
	err := r.babyDao.CreateBabyGrowthRecord(ctx, dao.CreateBabyGrowthRecordParams{
		RecordID:          rid,
		BabyID:            bid,
		RecordTime:        recordTime,
		Height:            pgtype.Float8{Float64: height, Valid: height != 0},
		Weight:            pgtype.Float8{Float64: weight, Valid: weight != 0},
		HeadCircumference: pgtype.Float8{Float64: headCircumference, Valid: headCircumference != 0},
		Remark:            pgtype.Text{String: remark, Valid: remark != ""},
		Ctime:             now,
		Utime:             now,
		CreatedBy:         uid,
	})
	if err != nil {
		r.log.Error(err)
		return "", ErrDefault
	}
	_ = cache.Del(ctx, r.rdb, cache.LatestGrowthKey(babyID))
	return rid.String(), nil
}

type CurvePoint struct {
	Time  int64
	Value float64
}

func (r *BabyRepo) ListHeightCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListHeightCurveByBabyIDBetween(ctx, dao.ListHeightCurveByBabyIDBetweenParams{
		BabyID:       bid,
		RecordTime:   from,
		RecordTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]CurvePoint, 0, len(rows))
	for _, v := range rows {
		items = append(items, CurvePoint{
			Time:  v.RecordTime,
			Value: v.Height.Float64,
		})
	}
	return items, nil
}

func (r *BabyRepo) ListWeightCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListWeightCurveByBabyIDBetween(ctx, dao.ListWeightCurveByBabyIDBetweenParams{
		BabyID:       bid,
		RecordTime:   from,
		RecordTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]CurvePoint, 0, len(rows))
	for _, v := range rows {
		items = append(items, CurvePoint{
			Time:  v.RecordTime,
			Value: v.Weight.Float64,
		})
	}
	return items, nil
}

func (r *BabyRepo) ListHeadCircumferenceCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListHeadCircumferenceCurveByBabyIDBetween(ctx, dao.ListHeadCircumferenceCurveByBabyIDBetweenParams{
		BabyID:       bid,
		RecordTime:   from,
		RecordTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]CurvePoint, 0, len(rows))
	for _, v := range rows {
		items = append(items, CurvePoint{
			Time:  v.RecordTime,
			Value: v.HeadCircumference.Float64,
		})
	}
	return items, nil
}
