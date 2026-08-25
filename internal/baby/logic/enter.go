package logic

import (
	"context"
	"errors"
	babyconstant "nurture/internal/baby/constant"
	"nurture/internal/baby/dto"
	"nurture/internal/baby/repo"
	"nurture/internal/pkg/zapx"
	"sort"
	"time"

	"github.com/google/uuid"
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

func (l *BabyLogic) SleepStart(ctx context.Context, userID string, uri dto.SleepStartUri) (dto.SleepStartResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.SleepStartResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.SleepStartResp{}, ErrDefault
	}
	sessionID, startedAt, e := l.babyRepo.StartSleep(ctx, uri.BabyID, userID)
	if e != nil {
		l.log.Error(e)
		return dto.SleepStartResp{}, ErrDefault
	}
	return dto.SleepStartResp{
		SessionID: sessionID,
		StartedAt: startedAt,
	}, nil
}

func (l *BabyLogic) SleepStop(ctx context.Context, userID string, uri dto.SleepStopUri, req dto.SleepStopReq) (dto.SleepStopResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.SleepStopResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.SleepStopResp{}, ErrDefault
	}
	sid, start, end, dur, e := l.babyRepo.StopSleep(ctx, req.SessionID)
	if e != nil {
		l.log.Error(e)
		return dto.SleepStopResp{}, ErrDefault
	}
	return dto.SleepStopResp{
		SessionID:  sid,
		StartedAt:  start,
		EndedAt:    end,
		DurationMs: dur,
	}, nil
}

func (l *BabyLogic) SleepActive(ctx context.Context, userID string, uri dto.SleepActiveUri) (dto.SleepActiveResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.SleepActiveResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.SleepActiveResp{}, ErrDefault
	}
	sid, startedAt, e := l.babyRepo.GetActiveSleep(ctx, uri.BabyID, userID)
	if e != nil {
		l.log.Error(e)
		return dto.SleepActiveResp{}, ErrDefault
	}
	if sid == "" {
		return dto.SleepActiveResp{}, nil
	}
	return dto.SleepActiveResp{
		SessionID: sid,
		StartedAt: startedAt,
	}, nil
}

func (l *BabyLogic) ListSleepAt(ctx context.Context, userID string, uri dto.SleepListAtUri, req dto.SleepListAtReq) (dto.SleepListAtResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.SleepListAtResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.SleepListAtResp{}, ErrDefault
	}
	if req.Date == "" {
		return dto.SleepListAtResp{}, ErrParamsType
	}
	t, parseErr := time.ParseInLocation("20060102", req.Date, time.UTC)
	if parseErr != nil {
		return dto.SleepListAtResp{}, ErrParamsType
	}
	from := t.UnixMilli()
	to := t.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	rows, e := l.babyRepo.ListSleepBetween(ctx, uri.BabyID, from, to)
	if e != nil {
		l.log.Error(e)
		return dto.SleepListAtResp{}, ErrDefault
	}
	var items []dto.SleepItem
	for _, v := range rows {
		items = append(items, dto.SleepItem{
			SessionID:  v.SessionID,
			StartedAt:  v.StartTime,
			EndedAt:    v.EndTime,
			DurationMs: v.Duration,
		})
	}
	return dto.SleepListAtResp{Items: items}, nil
}

func (l *BabyLogic) CreateFeeding(ctx context.Context, userID string, uri dto.FeedingCreateUri, req dto.FeedingCreateReq) (dto.FeedingCreateResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.FeedingCreateResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.FeedingCreateResp{}, ErrDefault
	}
	if req.FeedType != babyconstant.FeedBreastMilk && req.FeedType != babyconstant.FeedFormula && req.FeedType != babyconstant.FeedSolid {
		return dto.FeedingCreateResp{}, ErrParamsType
	}
	if req.FeedTime <= 0 {
		return dto.FeedingCreateResp{}, ErrParamsType
	}
	now := time.Now().UnixMilli()
	id, e := l.babyRepo.CreateFeeding(ctx, uri.BabyID, userID, req.FeedTime, req.FeedType, req.Remark, now)
	if e != nil {
		l.log.Error(e)
		return dto.FeedingCreateResp{}, ErrDefault
	}
	return dto.FeedingCreateResp{
		FeedingID: id,
		Message:   "创建成功",
	}, nil
}

func (l *BabyLogic) UpdateFeeding(ctx context.Context, userID string, uri dto.FeedingUpdateUri, req dto.FeedingUpdateReq) (dto.FeedingUpdateResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.FeedingUpdateResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.FeedingUpdateResp{}, ErrDefault
	}
	if req.FeedType != babyconstant.FeedBreastMilk && req.FeedType != babyconstant.FeedFormula && req.FeedType != babyconstant.FeedSolid {
		return dto.FeedingUpdateResp{}, ErrParamsType
	}
	if req.FeedTime <= 0 {
		return dto.FeedingUpdateResp{}, ErrParamsType
	}
	now := time.Now().UnixMilli()
	if e := l.babyRepo.UpdateFeeding(ctx, uri.BabyID, uri.FeedingID, req.FeedType, req.FeedTime, req.Remark, now); e != nil {
		l.log.Error(e)
		return dto.FeedingUpdateResp{}, ErrDefault
	}
	return dto.FeedingUpdateResp{
		Message: "更新成功",
	}, nil
}

func (l *BabyLogic) ListFeedingAt(ctx context.Context, userID string, uri dto.FeedingListAtUri, req dto.FeedingListAtReq) (dto.FeedingListAtResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.FeedingListAtResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.FeedingListAtResp{}, ErrDefault
	}
	if req.Date == "" {
		return dto.FeedingListAtResp{}, ErrParamsType
	}
	t, parseErr := time.ParseInLocation("20060102", req.Date, time.UTC)
	if parseErr != nil {
		return dto.FeedingListAtResp{}, ErrParamsType
	}
	from := t.UnixMilli()
	to := t.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	rows, e := l.babyRepo.ListFeedingBetween(ctx, uri.BabyID, from, to)
	if e != nil {
		l.log.Error(e)
		return dto.FeedingListAtResp{}, ErrDefault
	}
	var items []dto.FeedingItem
	for _, v := range rows {
		items = append(items, dto.FeedingItem{
			FeedingID: v.FeedingID,
			FeedTime:  v.FeedTime,
			FeedType:  v.FeedType,
			Remark:    v.Remark,
		})
	}
	return dto.FeedingListAtResp{Items: items}, nil
}

func (l *BabyLogic) CreateDiaper(ctx context.Context, userID string, uri dto.DiaperCreateUri, req dto.DiaperCreateReq) (dto.DiaperCreateResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.DiaperCreateResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.DiaperCreateResp{}, ErrDefault
	}
	if req.ChangeTime <= 0 {
		return dto.DiaperCreateResp{}, ErrParamsType
	}
	if !l.validateDiaper(req.DiaperType, req.PeeColor, req.PoopColor, req.PoopConsistency) {
		return dto.DiaperCreateResp{}, ErrParamsType
	}
	now := time.Now().UnixMilli()
	id, err := l.babyRepo.CreateDiaper(ctx, uri.BabyID, userID, req.ChangeTime, req.DiaperType, req.PeeColor, req.PoopColor, req.PoopConsistency, req.Remark, now)
	if err != nil {
		l.log.Error(err)
		return dto.DiaperCreateResp{}, ErrDefault
	}
	return dto.DiaperCreateResp{DiaperID: id, Message: "创建成功"}, nil
}

func (l *BabyLogic) UpdateDiaper(ctx context.Context, userID string, uri dto.DiaperUpdateUri, req dto.DiaperUpdateReq) (dto.DiaperUpdateResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.DiaperUpdateResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.DiaperUpdateResp{}, ErrDefault
	}
	if req.ChangeTime <= 0 {
		return dto.DiaperUpdateResp{}, ErrParamsType
	}
	if !l.validateDiaper(req.DiaperType, req.PeeColor, req.PoopColor, req.PoopConsistency) {
		return dto.DiaperUpdateResp{}, ErrParamsType
	}
	now := time.Now().UnixMilli()
	if err := l.babyRepo.UpdateDiaper(ctx, uri.BabyID, uri.DiaperID, req.DiaperType, req.ChangeTime, req.PeeColor, req.PoopColor, req.PoopConsistency, req.Remark, now); err != nil {
		l.log.Error(err)
		return dto.DiaperUpdateResp{}, ErrDefault
	}
	return dto.DiaperUpdateResp{Message: "更新成功"}, nil
}

func (l *BabyLogic) GetDiaperAt(ctx context.Context, userID string, uri dto.DiaperGetAtUri, req dto.DiaperGetAtReq) (dto.DiaperRecordResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.DiaperRecordResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.DiaperRecordResp{}, ErrDefault
	}
	if req.Date == "" {
		return dto.DiaperRecordResp{}, ErrParamsType
	}
	t, parseErr := time.ParseInLocation("20060102", req.Date, time.UTC)
	if parseErr != nil {
		return dto.DiaperRecordResp{}, ErrParamsType
	}
	from := t.UnixMilli()
	to := t.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	row, ok, e := l.babyRepo.GetDiaperBetween(ctx, uri.BabyID, from, to)
	if e != nil {
		l.log.Error(e)
		return dto.DiaperRecordResp{}, ErrDefault
	}
	if !ok {
		return dto.DiaperRecordResp{}, nil
	}
	return dto.DiaperRecordResp{
		DiaperID:        row.DiaperID,
		ChangeTime:      row.ChangeTime,
		DiaperType:      dto.EnumItem{ID: row.DiaperType, Name: l.diaperTypeName(row.DiaperType)},
		PeeColor:        l.toEnumPtr(row.PeeColor, l.peeColorName),
		PoopColor:       l.toEnumPtr(row.PoopColor, l.poopColorName),
		PoopConsistency: l.toEnumPtr(row.PoopConsistency, l.poopConsistencyName),
		Remark:          row.Remark,
	}, nil
}

func (l *BabyLogic) ListDiaperAt(ctx context.Context, userID string, uri dto.DiaperListAtUri, req dto.DiaperListAtReq) (dto.DiaperListAtResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.DiaperListAtResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.DiaperListAtResp{}, ErrDefault
	}
	if req.Date == "" {
		return dto.DiaperListAtResp{}, ErrParamsType
	}
	t, parseErr := time.ParseInLocation("20060102", req.Date, time.UTC)
	if parseErr != nil {
		return dto.DiaperListAtResp{}, ErrParamsType
	}
	from := t.UnixMilli()
	to := t.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	rows, e := l.babyRepo.ListDiaperBetween(ctx, uri.BabyID, from, to)
	if e != nil {
		l.log.Error(e)
		return dto.DiaperListAtResp{}, ErrDefault
	}
	var items []dto.DiaperItem
	for _, v := range rows {
		items = append(items, dto.DiaperItem{
			DiaperID:        v.DiaperID,
			ChangeTime:      v.ChangeTime,
			DiaperType:      dto.EnumItem{ID: v.DiaperType, Name: l.diaperTypeName(v.DiaperType)},
			PeeColor:        l.toEnumPtr(v.PeeColor, l.peeColorName),
			PoopColor:       l.toEnumPtr(v.PoopColor, l.poopColorName),
			PoopConsistency: l.toEnumPtr(v.PoopConsistency, l.poopConsistencyName),
			Remark:          v.Remark,
		})
	}
	return dto.DiaperListAtResp{Items: items}, nil
}

func (l *BabyLogic) DailyStats(ctx context.Context, userID string, uri dto.DailyStatsUri, req dto.DailyStatsReq) (dto.DailyStatsResp, error) {
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, uri.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return dto.DailyStatsResp{}, ErrBabyNotExist
		}
		l.log.Error(err)
		return dto.DailyStatsResp{}, ErrDefault
	}
	if req.Date == "" {
		return dto.DailyStatsResp{}, ErrParamsType
	}
	t, parseErr := time.ParseInLocation("20060102", req.Date, time.UTC)
	if parseErr != nil {
		return dto.DailyStatsResp{}, ErrParamsType
	}
	from := t.UnixMilli()
	to := t.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	s, e := l.babyRepo.GetDailyStats(ctx, uri.BabyID, from, to)
	if e != nil {
		l.log.Error(e)
		return dto.DailyStatsResp{}, ErrDefault
	}
	feedRows, fe := l.babyRepo.ListFeedingBetween(ctx, uri.BabyID, from, to)
	if fe != nil {
		l.log.Error(fe)
		return dto.DailyStatsResp{}, ErrDefault
	}
	diaperRows, de := l.babyRepo.ListDiaperBetween(ctx, uri.BabyID, from, to)
	if de != nil {
		l.log.Error(de)
		return dto.DailyStatsResp{}, ErrDefault
	}
	sleepRows, se := l.babyRepo.ListSleepBetween(ctx, uri.BabyID, from, to)
	if se != nil {
		l.log.Error(se)
		return dto.DailyStatsResp{}, ErrDefault
	}
	items := make([]dto.DailyRecordItem, 0, len(feedRows)+len(diaperRows)+len(sleepRows))
	for _, v := range feedRows {
		items = append(items, dto.DailyRecordItem{
			ID:      v.FeedingID,
			Type:    "feeding",
			SubType: v.FeedType,
			Time:    v.FeedTime,
		})
	}
	for _, v := range diaperRows {
		items = append(items, dto.DailyRecordItem{
			ID:      v.DiaperID,
			Type:    "diaper",
			SubType: v.DiaperType,
			Time:    v.ChangeTime,
		})
	}
	for _, v := range sleepRows {
		items = append(items, dto.DailyRecordItem{
			ID:         v.SessionID,
			Type:       "sleep",
			SubType:    "duration",
			Time:       v.StartTime,
			DurationMs: v.Duration,
		})
	}
	// 统一按时间升序
	sort.Slice(items, func(i, j int) bool { return items[i].Time < items[j].Time })
	return dto.DailyStatsResp{
		FeedingCount:    s.FeedingCount,
		SleepDurationMs: s.SleepDurationMs,
		DiaperCount:     s.DiaperCount,
		Items:           items,
	}, nil
}

func (l *BabyLogic) validateDiaper(diaperType, peeColor, poopColor, poopConsistency string) bool {
	switch diaperType {
	case "pee":
		if peeColor == "" || !l.isValidPeeColor(peeColor) {
			return false
		}
		return poopColor == "" && poopConsistency == ""
	case "poop":
		if poopColor == "" || poopConsistency == "" || !l.isValidPoopColor(poopColor) || !l.isValidPoopConsistency(poopConsistency) {
			return false
		}
		return peeColor == ""
	case "both":
		return peeColor != "" && poopColor != "" && poopConsistency != "" &&
			l.isValidPeeColor(peeColor) && l.isValidPoopColor(poopColor) && l.isValidPoopConsistency(poopConsistency)
	case "dry":
		return peeColor == "" && poopColor == "" && poopConsistency == ""
	default:
		return false
	}
}

func (l *BabyLogic) diaperTypeName(id string) string {
	switch id {
	case "pee":
		return "嘘嘘"
	case "poop":
		return "便便"
	case "both":
		return "嘘嘘+便便"
	case "dry":
		return "干爽"
	default:
		return id
	}
}

func (l *BabyLogic) peeColorName(id string) string {
	switch id {
	case "milky_white":
		return "乳白色"
	case "pink":
		return "粉色"
	case "normal":
		return "正常"
	case "yellow":
		return "黄色"
	case "red":
		return "红色"
	case "tea":
		return "浓茶色"
	default:
		return id
	}
}

func (l *BabyLogic) poopConsistencyName(id string) string {
	switch id {
	case "paste":
		return "膏状"
	case "foamy":
		return "泡沫样"
	case "milky":
		return "有奶瓣"
	case "food_residue":
		return "有食物残渣"
	case "egg_flower":
		return "蛋花样"
	case "watery":
		return "水样便"
	case "sheep":
		return "羊屎便"
	case "bloody":
		return "含血便"
	default:
		return id
	}
}

func (l *BabyLogic) poopColorName(id string) string {
	switch id {
	case "dark_green":
		return "墨绿色"
	case "green":
		return "绿色"
	case "yellow":
		return "黄色"
	case "orange":
		return "棕色"
	case "red":
		return "红色"
	case "black":
		return "黑色"
	case "gray_white":
		return "灰白色"
	default:
		return id
	}
}

func (l *BabyLogic) isValidPeeColor(id string) bool {
	switch id {
	case "milky_white", "pink", "normal", "yellow", "red", "tea":
		return true
	}
	return false
}
func (l *BabyLogic) isValidPoopColor(id string) bool {
	switch id {
	case "dark_green", "green", "yellow", "orange", "red", "black", "gray_white":
		return true
	}
	return false
}
func (l *BabyLogic) isValidPoopConsistency(id string) bool {
	switch id {
	case "paste", "foamy", "milky", "food_residue", "egg_flower", "watery", "sheep", "bloody":
		return true
	}
	return false
}

func (l *BabyLogic) toEnumPtr(id string, nameFn func(string) string) *dto.EnumItem {
	if id == "" {
		return nil
	}
	n := nameFn(id)
	return &dto.EnumItem{ID: id, Name: n}
}
func (l *BabyLogic) ChangeBaby(ctx context.Context, userID string, req dto.ChangeBabyReq) (dto.ChangeBabyResp, error) {
	var resp dto.ChangeBabyResp
	rows, err := l.babyRepo.ListMyBabies(ctx, userID)
	if err != nil {
		return resp, ErrDefault
	}
	items := make([]dto.BabyItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.BabyItem{
			BabyID: v.BabyID,
			Name:   v.Name,
			Avatar: v.Avatar,
		})
	}
	resp.Babies = items
	return resp, nil
}

func (l *BabyLogic) CreateGrowth(ctx context.Context, userID string, req dto.CreateGrowthReq) (dto.CreateGrowthResp, error) {
	var resp dto.CreateGrowthResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	h := req.Height
	w := req.Weight
	hc := req.HeadCircumference
	gr, e := l.babyRepo.GetLatestGrowthByBabyID(ctx, req.BabyID)
	if e != nil && !errors.Is(e, repo.ErrBabyGrowthNotExist) {
		l.log.Error(e)
		return resp, ErrDefault
	}
	if e == nil {
		if h == 0 && gr.Height.Valid {
			h = gr.Height.Float64
		}
		if w == 0 && gr.Weight.Valid {
			w = gr.Weight.Float64
		}
		if hc == 0 && gr.HeadCircumference.Valid {
			hc = gr.HeadCircumference.Float64
		}
		rtDay := time.UnixMilli(req.RecordTime).UTC().Truncate(24 * time.Hour).UnixMilli()
		grDay := time.UnixMilli(gr.RecordTime).UTC().Truncate(24 * time.Hour).UnixMilli()
		if rtDay == grDay {
			if err := l.babyRepo.UpdateGrowthByRecordID(ctx, gr.RecordID.String(), req.RecordTime, h, w, hc, req.Remark, userID); err != nil {
				l.log.Error(err)
				return resp, ErrDefault
			}
			resp.RecordID = gr.RecordID.String()
			resp.Message = "更新成功"
			return resp, nil
		}
	}
	rid, err := l.babyRepo.CreateGrowthRecord(ctx, req.BabyID, userID, req.RecordTime, h, w, hc, req.Remark)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.RecordID = rid
	resp.Message = "创建成功"
	return resp, nil
}

func (l *BabyLogic) NewBaby(ctx context.Context, userID string, req dto.NewBabyReq) (dto.NewBabyResp, error) {
	var resp dto.NewBabyResp
	babyID := uuid.NewString()
	now := time.Now().UnixMilli()
	var partnerID string
	if l.partnerReader != nil {
		var err error
		partnerID, err = l.partnerReader.GetPartnerByUserID(ctx, userID)
		if err != nil {
			l.log.Error(err)
			return resp, ErrDefault
		}
	}
	err := l.babyRepo.CreateBabyWithInit(ctx, userID, partnerID, babyID, req.Name, req.Gender, req.Birthday, req.Avatar,
		now, req.Height, req.Weight, req.HeadCircumference, req.Remark)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.BabyID = babyID
	resp.Message = "宝宝创建成功"
	return resp, nil
}

func (l *BabyLogic) GetProfile(ctx context.Context, userID string, req dto.BabyProfileReq) (dto.BabyProfileResp, error) {
	var resp dto.BabyProfileResp
	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	gr, err := l.babyRepo.GetLatestGrowthByBabyID(ctx, req.BabyID)
	if err != nil && !errors.Is(err, repo.ErrBabyGrowthNotExist) {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.BabyID = b.BabyID.String()
	resp.Name = b.Name
	resp.Avatar = b.Avatar
	resp.Gender = b.Gender
	resp.Birthday = b.Birthday
	if err == nil {
		resp.RecordTime = gr.RecordTime
		if gr.Height.Valid {
			resp.Height = gr.Height.Float64
		}
		if gr.Weight.Valid {
			resp.Weight = gr.Weight.Float64
		}
		if gr.HeadCircumference.Valid {
			resp.HeadCircumference = gr.HeadCircumference.Float64
		}
	}
	return resp, nil
}

func (l *BabyLogic) GetGrowthAt(ctx context.Context, userID string, req dto.GrowthAtReq) (dto.GrowthAtResp, error) {
	var resp dto.GrowthAtResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	dayStart := time.UnixMilli(req.Time).UTC().Truncate(24 * time.Hour)
	start := dayStart.UnixMilli()
	end := dayStart.Add(24 * time.Hour).Add(-time.Millisecond).UnixMilli()
	gr, e := l.babyRepo.GetGrowthByBabyIDBetween(ctx, req.BabyID, start, end)
	if e != nil {
		if errors.Is(e, repo.ErrBabyGrowthNotExist) {
			return resp, nil
		}
		l.log.Error(e)
		return resp, ErrDefault
	}
	resp.RecordID = gr.RecordID.String()
	resp.RecordTime = gr.RecordTime
	if gr.Height.Valid {
		v := gr.Height.Float64
		resp.Height = &v
	}
	if gr.Weight.Valid {
		v := gr.Weight.Float64
		resp.Weight = &v
	}
	if gr.HeadCircumference.Valid {
		v := gr.HeadCircumference.Float64
		resp.HeadCircumference = &v
	}
	if gr.Remark.Valid {
		resp.Remark = gr.Remark.String
	}
	if gr.CreatedBy.Valid {
		if uid, err := uuid.FromBytes(gr.CreatedBy.Bytes[:]); err == nil {
			resp.CreatedBy = uid.String()
		}
	}
	if gr.UpdatedBy.Valid {
		if uid, err := uuid.FromBytes(gr.UpdatedBy.Bytes[:]); err == nil {
			resp.UpdatedBy = uid.String()
		}
	}
	return resp, nil
}

func (l *BabyLogic) GrowthCurve(ctx context.Context, userID string, req dto.GrowthCurveReq) (dto.GrowthCurveResp, error) {
	var resp dto.GrowthCurveResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	to := req.To
	if to <= 0 || to > now {
		to = now
	}
	from := req.From
	if from > to {
		return resp, ErrParamsType
	}
	var unit string
	var rows []repo.CurvePoint
	switch req.Metric {
	case "height":
		unit = "cm"
		rows, err = l.babyRepo.ListHeightCurveBetween(ctx, req.BabyID, from, to)
	case "weight":
		unit = "kg"
		rows, err = l.babyRepo.ListWeightCurveBetween(ctx, req.BabyID, from, to)
	case "head_circumference":
		unit = "cm"
		rows, err = l.babyRepo.ListHeadCircumferenceCurveBetween(ctx, req.BabyID, from, to)
	default:
		return resp, ErrParamsType
	}
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	if req.GroupBy == "month" && len(rows) > 0 {
		var grouped []repo.CurvePoint
		curY := -1
		curM := time.Month(0)
		var last repo.CurvePoint
		for _, v := range rows {
			t := time.UnixMilli(v.Time).UTC()
			y, m, _ := t.Date()
			if curY == -1 {
				curY = y
				curM = m
			}
			if y != curY || m != curM {
				if last.Time != 0 {
					grouped = append(grouped, last)
				}
				curY = y
				curM = m
			}
			last = v
		}
		if last.Time != 0 {
			grouped = append(grouped, last)
		}
		rows = grouped
	}
	if req.MaxPoints > 0 && len(rows) > req.MaxPoints {
		n := len(rows)
		stride := (n + req.MaxPoints - 1) / req.MaxPoints
		var sampled []repo.CurvePoint
		for i := 0; i < n; i += stride {
			sampled = append(sampled, rows[i])
		}
		if sampled[len(sampled)-1].Time != rows[n-1].Time {
			sampled = append(sampled, rows[n-1])
		}
		rows = sampled
	}
	items := make([]dto.CurvePoint, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.CurvePoint{
			Time:  v.Time,
			Value: v.Value,
		})
	}
	resp.Metric = req.Metric
	resp.Unit = unit
	resp.Items = items
	return resp, nil
}
func (l *BabyLogic) GetVaccineList(ctx context.Context, userID string, req dto.GetVaccineListReq) (dto.GetVaccineListResp, error) {
	var resp dto.GetVaccineListResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	rows, err := l.babyRepo.ListVaccineRecordsByBaby(ctx, req.BabyID)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.VaccineItem, 0, len(rows))
	for _, v := range rows {
		if req.Status != "" && req.Status != "all" && v.Status != req.Status {
			continue
		}
		items = append(items, dto.VaccineItem{
			DoseID:     v.DoseID,
			VaccineID:  v.VaccineID,
			Name:       v.Name,
			Disease:    v.Disease,
			Link:       v.Link,
			DoseNumber: v.DoseNumber,
			DueTime:    v.DueTime,
			Status:     v.Status,
			ActualTime: v.ActualTime,
		})
	}
	resp.Items = items
	return resp, nil
}

func (l *BabyLogic) AdminCreateVaccine(ctx context.Context, req dto.AdminCreateVaccineReq) (dto.AdminCreateVaccineResp, error) {
	var resp dto.AdminCreateVaccineResp
	if req.Name == "" || req.Disease == "" || len(req.Doses) == 0 {
		return resp, ErrParamsType
	}
	vaccineID := uuid.NewString()
	doses := make([]repo.DoseSpec, 0, len(req.Doses))
	for _, d := range req.Doses {
		if d.DoseNumber <= 0 || d.RecommendAgeDays < 0 {
			return resp, ErrParamsType
		}
		doses = append(doses, repo.DoseSpec{
			DoseNumber:       d.DoseNumber,
			RecommendAgeDays: d.RecommendAgeDays,
		})
	}
	vid, created, err := l.babyRepo.AdminCreateVaccine(ctx, vaccineID, req.Name, req.Disease, req.Link, doses)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.VaccineID = vid
	resp.Doses = make([]dto.AdminCreatedDose, 0, len(created))
	for _, c := range created {
		resp.Doses = append(resp.Doses, dto.AdminCreatedDose{
			DoseID:           c.DoseID,
			DoseNumber:       c.DoseNumber,
			RecommendAgeDays: c.RecommendAgeDays,
		})
	}
	return resp, nil
}

func (l *BabyLogic) ChangeVaccineStatus(ctx context.Context, userID string, req dto.ChangeVaccineStatusReq) (dto.ChangeVaccineStatusResp, error) {
	var resp dto.ChangeVaccineStatusResp
	b, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	if req.Status == "given" {
		_, err = l.babyRepo.UpdateVaccineStatusGiven(ctx, b.BabyID.String(), req.DoseID, req.ActualTime, now)
	} else {
		_, err = l.babyRepo.UpdateVaccineStatusNotGiven(ctx, b.BabyID.String(), req.DoseID, now)
	}
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Message = "操作成功"
	return resp, nil
}

func (l *BabyLogic) UploadBabyPhotos(ctx context.Context, userID string, req dto.UploadBabyPhotosReq) (dto.UploadBabyPhotosResp, error) {
	var resp dto.UploadBabyPhotosResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	now := time.Now().UnixMilli()
	rows, err := l.babyRepo.UploadPhotos(ctx, req.BabyID, req.Links, now)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.PhotoItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.PhotoItem{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	resp.Items = items
	return resp, nil
}

func (l *BabyLogic) DeleteBabyPhotos(ctx context.Context, userID string, req dto.DeleteBabyPhotosReq) (dto.DeleteBabyPhotosResp, error) {
	var resp dto.DeleteBabyPhotosResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	n, err := l.babyRepo.DeletePhotos(ctx, req.BabyID, req.PhotoIDs)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	resp.Deleted = n
	return resp, nil
}

func (l *BabyLogic) ListBabyPhotos(ctx context.Context, userID string, req dto.ListBabyPhotosReq) (dto.ListBabyPhotosResp, error) {
	var resp dto.ListBabyPhotosResp
	_, err := l.babyRepo.GetBabyByIDAndUser(ctx, req.BabyID, userID)
	if err != nil {
		if errors.Is(err, repo.ErrBabyNotExist) {
			return resp, ErrBabyNotExist
		}
		l.log.Error(err)
		return resp, ErrDefault
	}
	rows, hasMore, err := l.babyRepo.ListPhotos(ctx, req.BabyID, req.Page, req.PageSize)
	if err != nil {
		l.log.Error(err)
		return resp, ErrDefault
	}
	items := make([]dto.PhotoItem, 0, len(rows))
	for _, v := range rows {
		items = append(items, dto.PhotoItem{
			PhotoID: v.PhotoID,
			Link:    v.Link,
			Ctime:   v.Ctime,
		})
	}
	resp.Items = items
	resp.HasMore = hasMore
	resp.Page = req.Page
	resp.PageSize = req.PageSize
	return resp, nil
}
