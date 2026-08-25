package handler

import (
	"nurture/internal/baby/dto"
	"nurture/internal/baby/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type BabyHandler struct {
	babyLogic logic.IBabyLogic
}

func NewBabyHandler(babyLogic logic.IBabyLogic) *BabyHandler {
	return &BabyHandler{
		babyLogic: babyLogic,
	}
}

func (h *BabyHandler) ChangeBaby(c *gin.Context) {
	cr := middleware.GetBind[dto.ChangeBabyReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ChangeBaby(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) NewBaby(c *gin.Context) {
	cr := middleware.GetBind[dto.NewBabyReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.NewBaby(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetProfile(c *gin.Context) {
	cr := middleware.GetBind[dto.BabyProfileReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetProfile(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetVaccineList(c *gin.Context) {
	cr := middleware.GetBind[dto.GetVaccineListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetVaccineList(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ChangeVaccineStatus(c *gin.Context) {
	cr := middleware.GetBind[dto.ChangeVaccineStatusReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ChangeVaccineStatus(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) AdminCreateVaccine(c *gin.Context) {
	cr := middleware.GetBind[dto.AdminCreateVaccineReq](c)
	resp, err := h.babyLogic.AdminCreateVaccine(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UploadBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.UploadBabyPhotosReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UploadBabyPhotos(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) DeleteBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.DeleteBabyPhotosReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.DeleteBabyPhotos(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.ListBabyPhotosReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListBabyPhotos(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateGrowth(c *gin.Context) {
	cr := middleware.GetBind[dto.CreateGrowthReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateGrowth(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetGrowthAt(c *gin.Context) {
	cr := middleware.GetBind[dto.GrowthAtReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetGrowthAt(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GrowthCurve(c *gin.Context) {
	cr := middleware.GetBind[dto.GrowthCurveReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GrowthCurve(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepStart(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepStartUri](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepStart(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepStop(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepStopUri](c)
	cr := middleware.GetBind[dto.SleepStopReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepStop(c.Request.Context(), userID, uri, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepActive(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepActiveUri](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepActive(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListSleepAt(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepListAtUri](c)
	q := middleware.GetBind[dto.SleepListAtReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListSleepAt(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateFeeding(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingCreateUri](c)
	cr := middleware.GetBind[dto.FeedingCreateReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateFeeding(c.Request.Context(), userID, uri, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UpdateFeeding(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingUpdateUri](c)
	cr := middleware.GetBind[dto.FeedingUpdateReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UpdateFeeding(c.Request.Context(), userID, uri, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListFeedingAt(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingListAtUri](c)
	q := middleware.GetBind[dto.FeedingListAtReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListFeedingAt(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateDiaper(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperCreateUri](c)
	cr := middleware.GetBind[dto.DiaperCreateReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateDiaper(c.Request.Context(), userID, uri, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UpdateDiaper(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperUpdateUri](c)
	cr := middleware.GetBind[dto.DiaperUpdateReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UpdateDiaper(c.Request.Context(), userID, uri, cr)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetDiaperAt(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperGetAtUri](c)
	q := middleware.GetBind[dto.DiaperGetAtReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetDiaperAt(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListDiaperAt(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperListAtUri](c)
	q := middleware.GetBind[dto.DiaperListAtReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListDiaperAt(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *BabyHandler) DailyStats(c *gin.Context) {
	uri := middleware.GetBind[dto.DailyStatsUri](c)
	q := middleware.GetBind[dto.DailyStatsReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.DailyStats(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}
