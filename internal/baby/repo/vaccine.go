package repo

import (
	"context"
	babyconstant "nurture/internal/baby/constant"
	"nurture/internal/baby/repo/cache"
	"nurture/internal/baby/repo/dao"

	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type DoseSpec struct {
	DoseNumber       int32
	RecommendAgeDays int32
}

type CreatedDose struct {
	DoseID           string
	DoseNumber       int32
	RecommendAgeDays int32
}

func (r *BabyRepo) AdminCreateVaccine(ctx context.Context, vaccineID, name, disease, link string, doses []DoseSpec) (string, []CreatedDose, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return "", nil, err
	}
	qtx := r.babyDao.WithTx(tx)
	var vid pgtype.UUID
	if err := vid.Scan(vaccineID); err != nil {
		_ = tx.Rollback(ctx)
		return "", nil, err
	}
	now := time.Now().UnixMilli()
	vret, err := qtx.CreateVaccine(ctx, dao.CreateVaccineParams{
		VaccineID: vid,
		Name:      name,
		Disease:   disease,
		Link:      link,
		Ctime:     now,
	})
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return "", nil, ErrDefault
	}
	created := make([]CreatedDose, 0, len(doses))
	for _, d := range doses {
		var did pgtype.UUID
		if err := did.Scan(uuid.NewString()); err != nil {
			_ = tx.Rollback(ctx)
			return "", nil, err
		}
		dr, err := qtx.CreateVaccineDose(ctx, dao.CreateVaccineDoseParams{
			DoseID:           did,
			VaccineID:        vid,
			DoseNumber:       d.DoseNumber,
			RecommendAgeDays: d.RecommendAgeDays,
			Ctime:            now,
		})
		if err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return "", nil, ErrDefault
		}
		if _, err := qtx.InitBabyVaccineRecordsForDose(ctx, dao.InitBabyVaccineRecordsForDoseParams{
			DoseID: did,
			Ctime:  now,
		}); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return "", nil, ErrDefault
		}
		created = append(created, CreatedDose{
			DoseID:           dr.DoseID,
			DoseNumber:       dr.DoseNumber,
			RecommendAgeDays: dr.RecommendAgeDays,
		})
	}
	if err := tx.Commit(ctx); err != nil {
		return "", nil, err
	}
	return vret, created, nil
}

func (r *BabyRepo) ListVaccineRecordsByBaby(ctx context.Context, babyID string) ([]dao.ListVaccineRecordsByBabyIDRow, error) {
	key := cache.VaccineListKey(babyID)
	{
		var cached []dao.ListVaccineRecordsByBabyIDRow
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached, nil
		}
	}
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListVaccineRecordsByBabyID(ctx, bid)
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	_ = cache.SetJSON(ctx, r.rdb, key, rows, time.Duration(babyconstant.VaccineListTTL)*time.Second)
	return rows, nil
}

func (r *BabyRepo) UpdateVaccineStatusGiven(ctx context.Context, babyID, doseID string, actualTime, utime int64) (int64, error) {
	var bid, did pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return 0, err
	}
	if err := did.Scan(doseID); err != nil {
		return 0, err
	}
	n, err := r.babyDao.UpdateVaccineStatusGiven(ctx, dao.UpdateVaccineStatusGivenParams{
		BabyID:     bid,
		DoseID:     did,
		ActualTime: pgtype.Int8{Int64: actualTime, Valid: true},
		Utime:      utime,
	})
	if err != nil {
		r.log.Error(err)
		return 0, ErrDefault
	}
	_ = cache.Del(ctx, r.rdb, cache.VaccineListKey(babyID))
	return n, nil
}

func (r *BabyRepo) UpdateVaccineStatusNotGiven(ctx context.Context, babyID, doseID string, utime int64) (int64, error) {
	var bid, did pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return 0, err
	}
	if err := did.Scan(doseID); err != nil {
		return 0, err
	}
	n, err := r.babyDao.UpdateVaccineStatusNotGiven(ctx, dao.UpdateVaccineStatusNotGivenParams{
		BabyID: bid,
		DoseID: did,
		Utime:  utime,
	})
	if err != nil {
		r.log.Error(err)
		return 0, ErrDefault
	}
	_ = cache.Del(ctx, r.rdb, cache.VaccineListKey(babyID))
	return n, nil
}
