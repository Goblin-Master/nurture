package repo

import (
	"context"
	"nurture/internal/global"
	"nurture/internal/repo/baby"
	"time"

	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type IBabyRepo interface {
	ListMyBabies(ctx context.Context, userID string) ([]baby.ListBabiesByUserIDRow, error)
	CreateBabyWithInit(ctx context.Context, userID, babyID, name, gender string, birthday int64, avatar string, recordTime int64, height, weight, headCircumference float64, remark string) error
	GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (baby.Baby, error)
	GetLatestGrowthByBabyIDAndUser(ctx context.Context, babyID, userID string) (baby.BabyGrowthRecord, error)
	ListVaccineRecordsByBaby(ctx context.Context, babyID string) ([]baby.ListVaccineRecordsByBabyIDRow, error)
	UpdateVaccineStatusGiven(ctx context.Context, babyID, doseID string, actualTime, utime int64) (int64, error)
	UpdateVaccineStatusNotGiven(ctx context.Context, babyID, doseID string, utime int64) (int64, error)
	UploadPhotos(ctx context.Context, babyID string, links []string, now int64) ([]PhotoRow, error)
	DeletePhotos(ctx context.Context, babyID string, photoIDs []string) (int64, error)
	ListPhotos(ctx context.Context, babyID string, page, pageSize int) ([]PhotoRow, bool, error)
}

type BabyRepo struct {
	babyDao *baby.Queries
}

func NewBabyRepo() *BabyRepo {
	return &BabyRepo{
		babyDao: baby.New(global.DB),
	}
}

var _ IBabyRepo = (*BabyRepo)(nil)

func (r *BabyRepo) ListMyBabies(ctx context.Context, userID string) ([]baby.ListBabiesByUserIDRow, error) {
	var uid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListBabiesByUserID(ctx, uid)
	if err != nil {
		global.Log.Error(err)
		return nil, ErrDefault
	}
	return rows, nil
}

func (r *BabyRepo) CreateBabyWithInit(ctx context.Context, userID, babyID, name, gender string, birthday int64, avatar string,
	recordTime int64, height, weight, headCircumference float64, remark string) error {
	tx, err := global.DB.Begin(ctx)
	if err != nil {
		return err
	}
	qtx := r.babyDao.WithTx(tx)
	var uid, bid pgtype.UUID
	if err := uid.Scan(userID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := bid.Scan(babyID); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	ctime := recordTime
	utime := recordTime
	if err := qtx.CreateBaby(ctx, baby.CreateBabyParams{
		BabyID:   bid,
		UserID:   uid,
		Name:     name,
		Gender:   gender,
		Birthday: birthday,
		Avatar:   avatar,
		Ctime:    ctime,
		Utime:    utime,
	}); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// 初始成长记录
	var rid pgtype.UUID
	if err := rid.Scan(uuid.NewString()); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := qtx.CreateBabyGrowthRecord(ctx, baby.CreateBabyGrowthRecordParams{
		RecordID:          rid,
		BabyID:            bid,
		UserID:            uid,
		RecordTime:        recordTime,
		Height:            pgtype.Float8{Float64: height, Valid: height != 0},
		Weight:            pgtype.Float8{Float64: weight, Valid: weight != 0},
		HeadCircumference: pgtype.Float8{Float64: headCircumference, Valid: headCircumference != 0},
		Remark:            pgtype.Text{String: remark, Valid: remark != ""},
		Ctime:             ctime,
		Utime:             utime,
	}); err != nil {
		global.Log.Error(err)
		_ = tx.Rollback(ctx)
		return ErrDefault
	}
	// 创建疫苗记录（按剂次）
	doses, err := qtx.ListAllDoses(ctx)
	if err != nil {
		global.Log.Error(err)
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
		if err := qtx.CreateBabyVaccineRecord(ctx, baby.CreateBabyVaccineRecordParams{
			RecordID: vrid,
			BabyID:   bid,
			DoseID:   d.DoseID,
			DueTime:  due,
			Ctime:    ctime,
			Utime:    utime,
		}); err != nil {
			global.Log.Error(err)
			_ = tx.Rollback(ctx)
			return ErrDefault
		}
	}
	_ = time.Now() // ensure time imported if not used otherwise
	return tx.Commit(ctx)
}

func (r *BabyRepo) GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (baby.Baby, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return baby.Baby{}, err
	}
	if err := uid.Scan(userID); err != nil {
		return baby.Baby{}, err
	}
	b, err := r.babyDao.GetBabyByIDAndUser(ctx, baby.GetBabyByIDAndUserParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return baby.Baby{}, ErrBabyNotExist
		}
		global.Log.Error(err)
		return baby.Baby{}, ErrDefault
	}
	return b, nil
}

func (r *BabyRepo) GetLatestGrowthByBabyIDAndUser(ctx context.Context, babyID, userID string) (baby.BabyGrowthRecord, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return baby.BabyGrowthRecord{}, err
	}
	if err := uid.Scan(userID); err != nil {
		return baby.BabyGrowthRecord{}, err
	}
	gr, err := r.babyDao.GetLatestGrowthByBabyIDAndUser(ctx, baby.GetLatestGrowthByBabyIDAndUserParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return baby.BabyGrowthRecord{}, ErrBabyGrowthNotExist
		}
		global.Log.Error(err)
		return baby.BabyGrowthRecord{}, ErrDefault
	}
	return gr, nil
}

func (r *BabyRepo) ListVaccineRecordsByBaby(ctx context.Context, babyID string) ([]baby.ListVaccineRecordsByBabyIDRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListVaccineRecordsByBabyID(ctx, bid)
	if err != nil {
		global.Log.Error(err)
		return nil, ErrDefault
	}
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
	n, err := r.babyDao.UpdateVaccineStatusGiven(ctx, baby.UpdateVaccineStatusGivenParams{
		BabyID:     bid,
		DoseID:     did,
		ActualTime: pgtype.Int8{Int64: actualTime, Valid: true},
		Utime:      utime,
	})
	if err != nil {
		global.Log.Error(err)
		return 0, ErrDefault
	}
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
	n, err := r.babyDao.UpdateVaccineStatusNotGiven(ctx, baby.UpdateVaccineStatusNotGivenParams{
		BabyID: bid,
		DoseID: did,
		Utime:  utime,
	})
	if err != nil {
		global.Log.Error(err)
		return 0, ErrDefault
	}
	return n, nil
}

type PhotoRow struct {
	PhotoID string
	Link    string
	Ctime   int64
}

func (r *BabyRepo) UploadPhotos(ctx context.Context, babyID string, links []string, now int64) ([]PhotoRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.UploadBabyPhotos(ctx, baby.UploadBabyPhotosParams{
		BabyID:  bid,
		Column2: links,
		Ctime:   now,
	})
	if err != nil {
		global.Log.Error(err)
		return nil, ErrDefault
	}
	var items []PhotoRow
	for _, v := range rows {
		items = append(items, PhotoRow{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	return items, nil
}

func (r *BabyRepo) DeletePhotos(ctx context.Context, babyID string, photoIDs []string) (int64, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return 0, err
	}
	ids := make([]pgtype.UUID, 0, len(photoIDs))
	for _, s := range photoIDs {
		var id pgtype.UUID
		if err := id.Scan(s); err != nil {
			return 0, err
		}
		ids = append(ids, id)
	}
	n, err := r.babyDao.DeleteBabyPhotos(ctx, baby.DeleteBabyPhotosParams{
		BabyID:  bid,
		Column2: ids,
	})
	if err != nil {
		global.Log.Error(err)
		return 0, ErrDefault
	}
	return n, nil
}

func (r *BabyRepo) ListPhotos(ctx context.Context, babyID string, page, pageSize int) ([]PhotoRow, bool, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, false, err
	}
	offset := (page - 1) * pageSize
	limit := pageSize + 1
	rows, err := r.babyDao.ListBabyPhotos(ctx, baby.ListBabyPhotosParams{
		BabyID: bid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		global.Log.Error(err)
		return nil, false, ErrDefault
	}
	var items []PhotoRow
	for _, v := range rows {
		items = append(items, PhotoRow{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	hasMore := false
	if len(items) > pageSize {
		hasMore = true
		items = items[:pageSize]
	}
	return items, hasMore, nil
}
