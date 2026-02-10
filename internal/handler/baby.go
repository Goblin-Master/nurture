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
