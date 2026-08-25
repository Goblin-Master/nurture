package repo

import (
	"context"
	"nurture/internal/baby/repo/dao"
	"nurture/internal/pkg/zapx"

	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type IBabyRepo interface {
	ListMyBabies(ctx context.Context, userID string) ([]dao.ListBabiesByUserIDRow, error)
	HandlePartnerBoundEvent(ctx context.Context, eventID, fatherUserID, motherUserID string) (bool, error)
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
