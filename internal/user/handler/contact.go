package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (uh *UserHandler) GetBindPhoneCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetSMSCodeReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.GetBindPhoneCode(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) BindPhone(c *gin.Context) {
	cr := middleware.GetBind[dto.BindPhoneReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.BindPhone(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetBindEmailCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	resp, err := uh.userLogic.GetBindEmailCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) BindEmail(c *gin.Context) {
	cr := middleware.GetBind[dto.BindEmailReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.BindEmail(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetRebindPhoneCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetSMSCodeReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.GetRebindPhoneCode(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) RebindPhone(c *gin.Context) {
	cr := middleware.GetBind[dto.BindPhoneReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.RebindPhone(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetRebindEmailCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.GetRebindEmailCode(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) RebindEmail(c *gin.Context) {
	cr := middleware.GetBind[dto.BindEmailReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.RebindEmail(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}
