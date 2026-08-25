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

func (r *BabyRepo) ListMyBabies(ctx context.Context, userID string) ([]dao.ListBabiesByUserIDRow, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListBabiesByUserID(ctx, uid)
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	return rows, nil
}

func (r *BabyRepo) SyncPartnerBabies(ctx context.Context, fatherUserID, motherUserID string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	qtx := r.babyDao.WithTx(tx)
	var fatherUUID, motherUUID pgtype.UUID
	if err := fatherUUID.Scan(fatherUserID); err != nil {
		return err
	}
	if err := motherUUID.Scan(motherUserID); err != nil {
		return err
	}

	if err := r.copyBabies(ctx, qtx, fatherUUID, motherUUID); err != nil {
		return err
	}
	if err := r.copyBabies(ctx, qtx, motherUUID, fatherUUID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *BabyRepo) copyBabies(ctx context.Context, qtx *dao.Queries, fromUserID, toUserID pgtype.UUID) error {
	rows, err := qtx.ListBabiesByUserID(ctx, fromUserID)
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	for _, row := range rows {
		var babyID pgtype.UUID
		if err := babyID.Scan(row.BabyID); err != nil {
			return err
		}
		full, err := qtx.GetBabyByIDAndUser(ctx, dao.GetBabyByIDAndUserParams{
			BabyID: babyID,
			UserID: fromUserID,
		})
		if err != nil {
			r.log.Error(err)
			return ErrDefault
		}
		now := time.Now().UnixMilli()
		if err := qtx.CreateBaby(ctx, dao.CreateBabyParams{
			BabyID:   full.BabyID,
			UserID:   toUserID,
			Name:     full.Name,
			Gender:   full.Gender,
			Birthday: full.Birthday,
			Avatar:   full.Avatar,
			Ctime:    now,
			Utime:    now,
		}); err != nil {
			r.log.Error(err)
			return ErrDefault
		}
	}
	return nil
}

func (r *BabyRepo) CreateBabyWithInit(ctx context.Context, userID, partnerID, babyID, name, gender string, birthday int64, avatar string,
	recordTime int64, height, weight, headCircumference float64, remark string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.babyDao.WithTx(tx)
	var uid, pid, bid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	hasPartner := partnerID != ""
	if hasPartner {
		if err := pid.Scan(partnerID); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
	}
	if err := bid.Scan(babyID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	ctime := recordTime
	utime := recordTime
	if err := qtx.CreateBaby(ctx, dao.CreateBabyParams{
		BabyID:   bid,
		UserID:   uid,
		Name:     name,
		Gender:   gender,
		Birthday: birthday,
		Avatar:   avatar,
		Ctime:    ctime,
		Utime:    utime,
	}); err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// 初始成长记录：仅当至少一个身体数据不为 0 时创建
	if height != 0 || weight != 0 || headCircumference != 0 {
		var rid pgtype.UUID
		if err := rid.Scan(uuid.NewString()); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		err := qtx.CreateBabyGrowthRecord(ctx, dao.CreateBabyGrowthRecordParams{
			RecordID:          rid,
			BabyID:            bid,
			RecordTime:        recordTime,
			Height:            pgtype.Float8{Float64: height, Valid: height != 0},
			Weight:            pgtype.Float8{Float64: weight, Valid: weight != 0},
			HeadCircumference: pgtype.Float8{Float64: headCircumference, Valid: headCircumference != 0},
			Remark:            pgtype.Text{String: remark, Valid: remark != ""},
			Ctime:             ctime,
			Utime:             utime,
			CreatedBy:         uid,
		})
		if err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	// 创建疫苗记录（按剂次）
	doses, err := qtx.ListAllDoses(ctx)
	if err != nil {
		r.log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	for _, d := range doses {
		due := birthday + int64(d.RecommendAgeDays)*24*3600*1000
		var vrid pgtype.UUID
		if err := vrid.Scan(uuid.NewString()); err != nil {
			_ = tx.Rollback(ctx)
			return err
		}
		if err := qtx.CreateBabyVaccineRecord(ctx, dao.CreateBabyVaccineRecordParams{
			RecordID: vrid,
			BabyID:   bid,
			DoseID:   d.DoseID,
			DueTime:  due,
			Ctime:    ctime,
			Utime:    utime,
		}); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	if hasPartner {
		if err := qtx.CreateBaby(ctx, dao.CreateBabyParams{
			BabyID:   bid,
			UserID:   pid,
			Name:     name,
			Gender:   gender,
			Birthday: birthday,
			Avatar:   avatar,
			Ctime:    ctime,
			Utime:    utime,
		}); err != nil {
			r.log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
		// 另一半的初始成长记录：同样仅当提供了身体数据时创建
		if height != 0 || weight != 0 || headCircumference != 0 {
			var rid2 pgtype.UUID
			if err := rid2.Scan(uuid.NewString()); err != nil {
				_ = tx.Rollback(ctx)
				return err
			}
			err := qtx.CreateBabyGrowthRecord(ctx, dao.CreateBabyGrowthRecordParams{
				RecordID:          rid2,
				BabyID:            bid,
				RecordTime:        recordTime,
				Height:            pgtype.Float8{Float64: height, Valid: height != 0},
				Weight:            pgtype.Float8{Float64: weight, Valid: weight != 0},
				HeadCircumference: pgtype.Float8{Float64: headCircumference, Valid: headCircumference != 0},
				Remark:            pgtype.Text{String: remark, Valid: remark != ""},
				Ctime:             ctime,
				Utime:             utime,
				CreatedBy:         pid,
			})
			if err != nil {
				r.log.Error(err)
				_ = tx.Rollback(ctx)
				return ErrDefault
			}
		}
	}
	return tx.Commit(ctx)
}

func (r *BabyRepo) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (dao.Baby, error) {
	key := cache.InfoKey(babyID, userID)
	{
		var cached dao.Baby
		if ok, _ := cache.GetJSON(ctx, r.rdb, key, &cached); ok {
			return cached, nil
		}
	}
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return dao.Baby{}, err
	}
	if err := uid.Scan(userID); err != nil {
		return dao.Baby{}, err
	}
	b, err := r.babyDao.GetBabyByIDAndUser(ctx, dao.GetBabyByIDAndUserParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return dao.Baby{}, ErrBabyNotExist
		}
		r.log.Error(err)
		return dao.Baby{}, ErrDefault
	}
	_ = cache.SetJSON(ctx, r.rdb, key, b, time.Duration(babyconstant.InfoTTL)*time.Second)
	return b, nil
}

func (r *BabyRepo) GetLatestGrowthByBabyIDAndUser(ctx context.Context, babyID, userID string) (dao.BabyGrowthRecord, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return dao.BabyGrowthRecord{}, err
	}
	return r.GetLatestGrowthByBabyID(ctx, babyID)
}
