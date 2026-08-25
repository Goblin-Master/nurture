package logic

import (
	"context"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"nurture/internal/pkg/zapx"

	"go.uber.org/zap"
)

type PartnerReader interface {
	GetPartnerByUserID(ctx context.Context, userID string) (string, error)
}

type IBabyLogic interface {
	ChangeBaby(ctx context.Context, userID string, req dto.ChangeBabyReq) (dto.ChangeBabyResp, error)
	NewBaby(ctx context.Context, userID string, req dto.NewBabyReq) (dto.NewBabyResp, error)
	GetProfile(ctx context.Context, userID string, req dto.BabyProfileReq) (dto.BabyProfileResp, error)
	GetVaccineList(ctx context.Context, userID string, req dto.GetVaccineListReq) (dto.GetVaccineListResp, error)
	ChangeVaccineStatus(ctx context.Context, userID string, req dto.ChangeVaccineStatusReq) (dto.ChangeVaccineStatusResp, error)
	// admin vaccine
	AdminCreateVaccine(ctx context.Context, req dto.AdminCreateVaccineReq) (dto.AdminCreateVaccineResp, error)
	UploadBabyPhotos(ctx context.Context, userID string, req dto.UploadBabyPhotosReq) (dto.UploadBabyPhotosResp, error)
	DeleteBabyPhotos(ctx context.Context, userID string, req dto.DeleteBabyPhotosReq) (dto.DeleteBabyPhotosResp, error)
	ListBabyPhotos(ctx context.Context, userID string, req dto.ListBabyPhotosReq) (dto.ListBabyPhotosResp, error)
	CreateGrowth(ctx context.Context, userID string, req dto.CreateGrowthReq) (dto.CreateGrowthResp, error)
	GetGrowthAt(ctx context.Context, userID string, req dto.GrowthAtReq) (dto.GrowthAtResp, error)
	GrowthCurve(ctx context.Context, userID string, req dto.GrowthCurveReq) (dto.GrowthCurveResp, error)
	DailyStats(ctx context.Context, userID string, uri dto.DailyStatsUri, req dto.DailyStatsReq) (dto.DailyStatsResp, error)
	// daily sleep
	SleepStart(ctx context.Context, userID string, uri dto.SleepStartUri) (dto.SleepStartResp, error)
	SleepStop(ctx context.Context, userID string, uri dto.SleepStopUri, req dto.SleepStopReq) (dto.SleepStopResp, error)
	SleepActive(ctx context.Context, userID string, uri dto.SleepActiveUri) (dto.SleepActiveResp, error)
	ListSleepAt(ctx context.Context, userID string, uri dto.SleepListAtUri, req dto.SleepListAtReq) (dto.SleepListAtResp, error)
	// daily feeding
	CreateFeeding(ctx context.Context, userID string, uri dto.FeedingCreateUri, req dto.FeedingCreateReq) (dto.FeedingCreateResp, error)
	UpdateFeeding(ctx context.Context, userID string, uri dto.FeedingUpdateUri, req dto.FeedingUpdateReq) (dto.FeedingUpdateResp, error)
	ListFeedingAt(ctx context.Context, userID string, uri dto.FeedingListAtUri, req dto.FeedingListAtReq) (dto.FeedingListAtResp, error)
	// daily diaper
	CreateDiaper(ctx context.Context, userID string, uri dto.DiaperCreateUri, req dto.DiaperCreateReq) (dto.DiaperCreateResp, error)
	UpdateDiaper(ctx context.Context, userID string, uri dto.DiaperUpdateUri, req dto.DiaperUpdateReq) (dto.DiaperUpdateResp, error)
	GetDiaperAt(ctx context.Context, userID string, uri dto.DiaperGetAtUri, req dto.DiaperGetAtReq) (dto.DiaperRecordResp, error)
	ListDiaperAt(ctx context.Context, userID string, uri dto.DiaperListAtUri, req dto.DiaperListAtReq) (dto.DiaperListAtResp, error)
}

type BabyLogic struct {
	babyRepo      repo.IBabyRepo
	partnerReader PartnerReader
	log           *zap.SugaredLogger
}

func NewBabyLogic(babyRepo repo.IBabyRepo, partnerReader PartnerReader, log *zap.SugaredLogger) *BabyLogic {
	return &BabyLogic{
		babyRepo:      babyRepo,
		partnerReader: partnerReader,
		log:           zapx.OrNop(log),
	}
}

var _ IBabyLogic = (*BabyLogic)(nil)
