package repo

import (
	"context"
	babyconstant "nurture/internal/baby/constant"
	"nurture/internal/baby/repo/cache"
	"nurture/internal/baby/repo/dao"
	"nurture/internal/pkg/zapx"

	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type IBabyRepo interface {
	ListMyBabies(ctx context.Context, userID string) ([]dao.ListBabiesByUserIDRow, error)
	SyncPartnerBabies(ctx context.Context, fatherUserID, motherUserID string) error
	CreateBabyWithInit(ctx context.Context, userID, partnerID, babyID, name, gender string, birthday int64, avatar string, recordTime int64, height, weight, headCircumference float64, remark string) error
	GetBabyByIDAndUser(ctx context.Context, babyID, userID string) (dao.Baby, error)
	GetLatestGrowthByBabyIDAndUser(ctx context.Context, babyID, userID string) (dao.BabyGrowthRecord, error)
	GetLatestGrowthByBabyID(ctx context.Context, babyID string) (dao.BabyGrowthRecord, error)
	GetGrowthByBabyIDBetween(ctx context.Context, babyID string, start, end int64) (dao.BabyGrowthRecord, error)
	ListGrowthRecordsByBabyIDBetween(ctx context.Context, babyID string, start, end int64) ([]dao.BabyGrowthRecord, error)
	UpdateGrowthByRecordID(ctx context.Context, recordID string, recordTime int64, height, weight, headCircumference float64, remark string, updatedBy string) error
	CreateGrowthRecord(ctx context.Context, babyID, userID string, recordTime int64, height, weight, headCircumference float64, remark string) (string, error)
	ListVaccineRecordsByBaby(ctx context.Context, babyID string) ([]dao.ListVaccineRecordsByBabyIDRow, error)
	UpdateVaccineStatusGiven(ctx context.Context, babyID, doseID string, actualTime, utime int64) (int64, error)
	UpdateVaccineStatusNotGiven(ctx context.Context, babyID, doseID string, utime int64) (int64, error)
	UploadPhotos(ctx context.Context, babyID string, links []string, now int64) ([]PhotoRow, error)
	DeletePhotos(ctx context.Context, babyID string, photoIDs []string) (int64, error)
	ListPhotos(ctx context.Context, babyID string, page, pageSize int) ([]PhotoRow, bool, error)
	ListHeightCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error)
	ListWeightCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error)
	ListHeadCircumferenceCurveBetween(ctx context.Context, babyID string, from, to int64) ([]CurvePoint, error)
	StartSleep(ctx context.Context, babyID, userID string) (string, int64, error)
	StopSleep(ctx context.Context, sessionID string) (string, int64, int64, int64, error)
	GetActiveSleep(ctx context.Context, babyID, userID string) (string, int64, error)
	ListSleepBetween(ctx context.Context, babyID string, from, to int64) ([]SleepRow, error)
	CreateFeeding(ctx context.Context, babyID, userID string, feedTime int64, feedType, remark string, now int64) (string, error)
	UpdateFeeding(ctx context.Context, babyID, feedingID string, feedType string, feedTime int64, remark string, now int64) error
	ListFeedingBetween(ctx context.Context, babyID string, from, to int64) ([]FeedingRow, error)
	// diaper
	CreateDiaper(ctx context.Context, babyID, userID string, changeTime int64, diaperType, peeColor, poopColor, poopConsistency, remark string, now int64) (string, error)
	UpdateDiaper(ctx context.Context, babyID, diaperID string, diaperType string, changeTime int64, peeColor, poopColor, poopConsistency, remark string, now int64) error
	GetDiaperBetween(ctx context.Context, babyID string, from, to int64) (DiaperRow, bool, error)
	ListDiaperBetween(ctx context.Context, babyID string, from, to int64) ([]DiaperRow, error)
	GetDailyStats(ctx context.Context, babyID string, from, to int64) (DailyStats, error)
	AdminCreateVaccine(ctx context.Context, vaccineID, name, disease, link string, doses []DoseSpec) (string, []CreatedDose, error)
}

type BabyRepo struct {
	db      *pgxpool.Pool
	babyDao *dao.Queries
	rdb     redis.Cmdable
	log     *zap.SugaredLogger
}

func NewBabyRepo(db *pgxpool.Pool, rdb redis.Cmdable, log *zap.SugaredLogger) *BabyRepo {
	return &BabyRepo{
		db:      db,
		babyDao: dao.New(db),
		rdb:     rdb,
		log:     zapx.OrNop(log),
	}
}

var _ IBabyRepo = (*BabyRepo)(nil)

type DoseSpec struct {
	DoseNumber       int32
	RecommendAgeDays int32
}

type CreatedDose struct {
	DoseID           string
	DoseNumber       int32
	RecommendAgeDays int32
}

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

// 管理员：创建疫苗 + 多个剂次，并为所有宝宝初始化接种记录
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

type PhotoRow struct {
	PhotoID string
	Link    string
	Ctime   int64
}

type CurvePoint struct {
	Time  int64
	Value float64
}

type FeedingRow struct {
	FeedingID string
	FeedTime  int64
	FeedType  string
	Remark    string
}

type DiaperRow struct {
	DiaperID        string
	ChangeTime      int64
	DiaperType      string
	PeeColor        string
	PoopColor       string
	PoopConsistency string
	Remark          string
}

type DailyStats struct {
	FeedingCount    int64
	SleepDurationMs int64
	DiaperCount     int64
}

func (r *BabyRepo) UploadPhotos(ctx context.Context, babyID string, links []string, now int64) ([]PhotoRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.UploadBabyPhotos(ctx, dao.UploadBabyPhotosParams{
		BabyID:  bid,
		Column2: links,
		Ctime:   now,
	})
	if err != nil {
		r.log.Error(err)
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
	n, err := r.babyDao.DeleteBabyPhotos(ctx, dao.DeleteBabyPhotosParams{
		BabyID:  bid,
		Column2: ids,
	})
	if err != nil {
		r.log.Error(err)
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
	rows, err := r.babyDao.ListBabyPhotos(ctx, dao.ListBabyPhotosParams{
		BabyID: bid,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		r.log.Error(err)
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

func (r *BabyRepo) StartSleep(ctx context.Context, babyID, userID string) (string, int64, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", 0, err
	}
	if err := uid.Scan(userID); err != nil {
		return "", 0, err
	}
	row, err := r.babyDao.StartSleep(ctx, dao.StartSleepParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		r.log.Error(err)
		return "", 0, ErrDefault
	}
	return row.SleepID, row.StartTime, nil
}

func (r *BabyRepo) StopSleep(ctx context.Context, sessionID string) (string, int64, int64, int64, error) {
	var sid pgtype.UUID
	if err := sid.Scan(sessionID); err != nil {
		return "", 0, 0, 0, err
	}
	row, err := r.babyDao.StopSleep(ctx, sid)
	if err != nil {
		r.log.Error(err)
		return "", 0, 0, 0, ErrDefault
	}
	return row.SleepID, row.StartTime, row.EndTime.Int64, row.Duration.Int64, nil
}

func (r *BabyRepo) ForceStopSleepWithCap(ctx context.Context, sessionID string, capMs int64) (string, int64, int64, int64, error) {
	var sid pgtype.UUID
	if err := sid.Scan(sessionID); err != nil {
		return "", 0, 0, 0, err
	}
	row, err := r.babyDao.ForceStopSleepWithCap(ctx, dao.ForceStopSleepWithCapParams{
		SleepID:  sid,
		Duration: pgtype.Int8{Int64: capMs, Valid: true},
	})
	if err != nil {
		r.log.Error(err)
		return "", 0, 0, 0, ErrDefault
	}
	return row.SleepID, row.StartTime, row.EndTime.Int64, row.Duration.Int64, nil
}

func (r *BabyRepo) GetActiveSleep(ctx context.Context, babyID, userID string) (string, int64, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", 0, err
	}
	if err := uid.Scan(userID); err != nil {
		return "", 0, err
	}
	row, err := r.babyDao.GetActiveSleep(ctx, dao.GetActiveSleepParams{
		BabyID: bid,
		UserID: uid,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", 0, nil
		}
		r.log.Error(err)
		return "", 0, ErrDefault
	}
	return row.SleepID, row.StartTime, nil
}

type SleepRow struct {
	SessionID string
	StartTime int64
	EndTime   int64
	Duration  int64
}

func (r *BabyRepo) ListSleepBetween(ctx context.Context, babyID string, from, to int64) ([]SleepRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListSleepByBabyBetween(ctx, dao.ListSleepByBabyBetweenParams{
		BabyID:    bid,
		EndTime:   pgtype.Int8{Int64: from, Valid: true},
		StartTime: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]SleepRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, SleepRow{
			SessionID: v.SessionID,
			StartTime: v.StartTime,
			EndTime:   v.EndTime.Int64,
			Duration:  v.Duration.Int64,
		})
	}
	return items, nil
}

func (r *BabyRepo) CreateFeeding(ctx context.Context, babyID, userID string, feedTime int64, feedType, remark string, now int64) (string, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", err
	}
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := r.babyDao.CreateFeeding(ctx, dao.CreateFeedingParams{
		BabyID:   bid,
		UserID:   uid,
		FeedTime: feedTime,
		FeedType: feedType,
		Column5:  remark,
		Ctime:    now,
	})
	if err != nil {
		r.log.Error(err)
		return "", ErrDefault
	}
	return row.FeedingID, nil
}

func (r *BabyRepo) UpdateFeeding(ctx context.Context, babyID, feedingID string, feedType string, feedTime int64, remark string, now int64) error {
	var bid, fid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return err
	}
	if err := fid.Scan(feedingID); err != nil {
		return err
	}
	_, err := r.babyDao.UpdateFeeding(ctx, dao.UpdateFeedingParams{
		BabyID:    bid,
		FeedingID: fid,
		FeedType:  feedType,
		FeedTime:  feedTime,
		Column5:   remark,
		Utime:     now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *BabyRepo) ListFeedingBetween(ctx context.Context, babyID string, from, to int64) ([]FeedingRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListFeedingByBabyBetween(ctx, dao.ListFeedingByBabyBetweenParams{
		BabyID:     bid,
		FeedTime:   from,
		FeedTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]FeedingRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, FeedingRow{
			FeedingID: v.FeedingID,
			FeedTime:  v.FeedTime,
			FeedType:  v.FeedType,
			Remark:    v.Remark,
		})
	}
	return items, nil
}

func (r *BabyRepo) CreateDiaper(ctx context.Context, babyID, userID string, changeTime int64, diaperType, peeColor, poopColor, poopConsistency, remark string, now int64) (string, error) {
	var bid, uid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return "", err
	}
	if err := uid.Scan(userID); err != nil {
		return "", err
	}
	row, err := r.babyDao.CreateDiaper(ctx, dao.CreateDiaperParams{
		BabyID:     bid,
		UserID:     uid,
		ChangeTime: changeTime,
		DiaperType: diaperType,
		Column5:    peeColor,
		Column6:    poopColor,
		Column7:    poopConsistency,
		Column8:    remark,
		Ctime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return "", ErrDefault
	}
	return row.DiaperID, nil
}

func (r *BabyRepo) UpdateDiaper(ctx context.Context, babyID, diaperID string, diaperType string, changeTime int64, peeColor, poopColor, poopConsistency, remark string, now int64) error {
	var bid, did pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return err
	}
	if err := did.Scan(diaperID); err != nil {
		return err
	}
	_, err := r.babyDao.UpdateDiaper(ctx, dao.UpdateDiaperParams{
		BabyID:     bid,
		DiaperID:   did,
		DiaperType: diaperType,
		ChangeTime: changeTime,
		Column5:    peeColor,
		Column6:    poopColor,
		Column7:    poopConsistency,
		Column8:    remark,
		Utime:      now,
	})
	if err != nil {
		r.log.Error(err)
		return ErrDefault
	}
	return nil
}

func (r *BabyRepo) GetDiaperBetween(ctx context.Context, babyID string, from, to int64) (DiaperRow, bool, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return DiaperRow{}, false, err
	}
	row, err := r.babyDao.GetDiaperByBabyBetween(ctx, dao.GetDiaperByBabyBetweenParams{
		BabyID:       bid,
		ChangeTime:   from,
		ChangeTime_2: to,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DiaperRow{}, false, nil
		}
		r.log.Error(err)
		return DiaperRow{}, false, ErrDefault
	}
	return DiaperRow{
		DiaperID:        row.DiaperID,
		ChangeTime:      row.ChangeTime,
		DiaperType:      row.DiaperType,
		PeeColor:        row.PeeColor,
		PoopColor:       row.PoopColor,
		PoopConsistency: row.PoopConsistency,
		Remark:          row.Remark,
	}, true, nil
}

func (r *BabyRepo) ListDiaperBetween(ctx context.Context, babyID string, from, to int64) ([]DiaperRow, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return nil, err
	}
	rows, err := r.babyDao.ListDiaperByBabyBetween(ctx, dao.ListDiaperByBabyBetweenParams{
		BabyID:       bid,
		ChangeTime:   from,
		ChangeTime_2: to,
	})
	if err != nil {
		r.log.Error(err)
		return nil, ErrDefault
	}
	items := make([]DiaperRow, 0, len(rows))
	for _, v := range rows {
		items = append(items, DiaperRow{
			DiaperID:        v.DiaperID,
			ChangeTime:      v.ChangeTime,
			DiaperType:      v.DiaperType,
			PeeColor:        v.PeeColor,
			PoopColor:       v.PoopColor,
			PoopConsistency: v.PoopConsistency,
			Remark:          v.Remark,
		})
	}
	return items, nil
}

func (r *BabyRepo) GetDailyStats(ctx context.Context, babyID string, from, to int64) (DailyStats, error) {
	var bid pgtype.UUID
	if err := bid.Scan(babyID); err != nil {
		return DailyStats{}, err
	}
	row, err := r.babyDao.GetDailyStatsByBaby(ctx, dao.GetDailyStatsByBabyParams{
		BabyID:     bid,
		FeedTime:   from,
		FeedTime_2: to,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return DailyStats{}, nil
		}
		r.log.Error(err)
		return DailyStats{}, ErrDefault
	}
	return DailyStats{
		FeedingCount:    row.FeedingCount,
		SleepDurationMs: row.SleepDurationMs,
		DiaperCount:     row.DiaperCount,
	}, nil
}
