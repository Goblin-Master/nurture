package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (uh *UserHandler) Login(c *gin.Context) {
	cr := middleware.GetBind[dto.LoginReq](c)
	resp, err := uh.userLogic.Login(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) RefreshToken(c *gin.Context) {
	cr := middleware.GetBind[dto.RefreshTokenReq](c)
	resp, err := uh.userLogic.RefreshToken(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Logout(c *gin.Context) {
	cr := middleware.GetBind[dto.LogoutReq](c)
	atoken, err := jwtx.BearerToken(c)
	if err != nil {
		response.Response(c, nil, err)
		return
	}
	resp, err := uh.userLogic.Logout(c.Request.Context(), atoken, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Register(c *gin.Context) {
	cr := middleware.GetBind[dto.RegisterReq](c)
	resp, err := uh.userLogic.Register(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) RegisterSMS(c *gin.Context) {
	cr := middleware.GetBind[dto.RegisterSMSReq](c)
	resp, err := uh.userLogic.RegisterSMS(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ResetPassword(c *gin.Context) {
	cr := middleware.GetBind[dto.ResetPasswordReq](c)
	resp, err := uh.userLogic.ResetPassword(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetLoginCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	resp, err := uh.userLogic.GetLoginCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetRegisterCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	resp, err := uh.userLogic.GetRegisterCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetRegisterSMSCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetSMSCodeReq](c)
	resp, err := uh.userLogic.GetRegisterSMSCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetResetCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	resp, err := uh.userLogic.GetResetCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}
