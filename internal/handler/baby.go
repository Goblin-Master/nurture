package handler

import (
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type BabyHandler struct {
	babyLogic *logic.BabyLogic
}

func NewBabyHandler() *BabyHandler {
	return &BabyHandler{
		babyLogic: logic.NewBabyLogic(),
	}
}

func (h *BabyHandler) ChangeBaby(c *gin.Context) {
	cr := middleware.GetBind[dto.ChangeBabyReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ChangeBaby(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) NewBaby(c *gin.Context) {
	cr := middleware.GetBind[dto.NewBabyReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.NewBaby(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetProfile(c *gin.Context) {
	cr := middleware.GetBind[dto.BabyProfileReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetProfile(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetVaccineList(c *gin.Context) {
	cr := middleware.GetBind[dto.GetVaccineListReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetVaccineList(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ChangeVaccineStatus(c *gin.Context) {
	cr := middleware.GetBind[dto.ChangeVaccineStatusReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ChangeVaccineStatus(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) AdminCreateVaccine(c *gin.Context) {
	cr := middleware.GetBind[dto.AdminCreateVaccineReq](c)
	global.Log.Info(cr)
	resp, err := h.babyLogic.AdminCreateVaccine(c.Request.Context(), cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UploadBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.UploadBabyPhotosReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UploadBabyPhotos(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) DeleteBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.DeleteBabyPhotosReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.DeleteBabyPhotos(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListBabyPhotos(c *gin.Context) {
	cr := middleware.GetBind[dto.ListBabyPhotosReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListBabyPhotos(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateGrowth(c *gin.Context) {
	cr := middleware.GetBind[dto.CreateGrowthReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateGrowth(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetGrowthAt(c *gin.Context) {
	cr := middleware.GetBind[dto.GrowthAtReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetGrowthAt(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GrowthCurve(c *gin.Context) {
	cr := middleware.GetBind[dto.GrowthCurveReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GrowthCurve(c.Request.Context(), userID, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepStart(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepStartUri](c)
	global.Log.Info(uri)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepStart(c.Request.Context(), userID, uri)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepStop(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepStopUri](c)
	cr := middleware.GetBind[dto.SleepStopReq](c)
	global.Log.Info(uri)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepStop(c.Request.Context(), userID, uri, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) SleepActive(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepActiveUri](c)
	global.Log.Info(uri)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.SleepActive(c.Request.Context(), userID, uri)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListSleepAt(c *gin.Context) {
	uri := middleware.GetBind[dto.SleepListAtUri](c)
	q := middleware.GetBind[dto.SleepListAtReq](c)
	global.Log.Info(uri)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListSleepAt(c.Request.Context(), userID, uri, q)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateFeeding(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingCreateUri](c)
	cr := middleware.GetBind[dto.FeedingCreateReq](c)
	global.Log.Info(uri)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateFeeding(c.Request.Context(), userID, uri, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UpdateFeeding(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingUpdateUri](c)
	cr := middleware.GetBind[dto.FeedingUpdateReq](c)
	global.Log.Info(uri)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UpdateFeeding(c.Request.Context(), userID, uri, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListFeedingAt(c *gin.Context) {
	uri := middleware.GetBind[dto.FeedingListAtUri](c)
	q := middleware.GetBind[dto.FeedingListAtReq](c)
	global.Log.Info(uri)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListFeedingAt(c.Request.Context(), userID, uri, q)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) CreateDiaper(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperCreateUri](c)
	cr := middleware.GetBind[dto.DiaperCreateReq](c)
	global.Log.Info(uri)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.CreateDiaper(c.Request.Context(), userID, uri, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) UpdateDiaper(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperUpdateUri](c)
	cr := middleware.GetBind[dto.DiaperUpdateReq](c)
	global.Log.Info(uri)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.UpdateDiaper(c.Request.Context(), userID, uri, cr)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) GetDiaperAt(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperGetAtUri](c)
	q := middleware.GetBind[dto.DiaperGetAtReq](c)
	global.Log.Info(uri)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.GetDiaperAt(c.Request.Context(), userID, uri, q)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) ListDiaperAt(c *gin.Context) {
	uri := middleware.GetBind[dto.DiaperListAtUri](c)
	q := middleware.GetBind[dto.DiaperListAtReq](c)
	global.Log.Info(uri)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.ListDiaperAt(c.Request.Context(), userID, uri, q)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}

func (h *BabyHandler) DailyStats(c *gin.Context) {
	uri := middleware.GetBind[dto.DailyStatsUri](c)
	q := middleware.GetBind[dto.DailyStatsReq](c)
	global.Log.Info(uri)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := h.babyLogic.DailyStats(c.Request.Context(), userID, uri, q)
	global.Log.Info(resp)
	response.Response(c, resp, err)
}
