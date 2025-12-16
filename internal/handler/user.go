package handler

import (
	"nurture/internal/dto"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/middleware"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userLogic *logic.UserLogic
}

func NewUserHandler() *UserHandler {
	return &UserHandler{
		userLogic: logic.NewUserLogic(),
	}
}

func (uh *UserHandler) Login(c *gin.Context) {
	cr := middleware.GetBind[dto.LoginReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.Login(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Register(c *gin.Context) {
	cr := middleware.GetBind[dto.RegisterReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.Register(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ResetPassword(c *gin.Context) {
	cr := middleware.GetBind[dto.ResetPasswordReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.ResetPassword(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetLoginCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.GetLoginCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetRegisterCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.GetRegisterCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetResetCode(c *gin.Context) {
	cr := middleware.GetBind[dto.GetCodeReq](c)
	global.Log.Info(cr)
	resp, err := uh.userLogic.GetResetCode(c.Request.Context(), cr)
	response.Response(c, resp, err)
}
