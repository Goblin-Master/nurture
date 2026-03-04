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

func (uh *UserHandler) UpdateProfile(c *gin.Context) {
	cr := middleware.GetBind[dto.UpdateUserAdditionReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.UpdateProfile(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) UpdateAvatar(c *gin.Context) {
	cr := middleware.GetBind[dto.UpdateAvatarReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.UpdateAvatar(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) BindPartner(c *gin.Context) {
	cr := middleware.GetBind[dto.PartnerBindReq](c)
	global.Log.Info(cr)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.BindPartner(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetPartner(c *gin.Context) {
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.GetPartner(c.Request.Context(), userID)
	response.Response(c, resp, err)
}

func (uh *UserHandler) MyProfile(c *gin.Context) {
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.MyProfile(c.Request.Context(), userID)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Follow(c *gin.Context) {
	uri := middleware.GetBind[dto.FollowReq](c)
	global.Log.Info(uri)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.Follow(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Unfollow(c *gin.Context) {
	uri := middleware.GetBind[dto.FollowReq](c)
	global.Log.Info(uri)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.Unfollow(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ListFollowing(c *gin.Context) {
	q := middleware.GetBind[dto.FollowingListReq](c)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.ListFollowing(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ListFollowers(c *gin.Context) {
	q := middleware.GetBind[dto.FollowersListReq](c)
	global.Log.Info(q)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.ListFollowers(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

// admin
func (uh *UserHandler) AdminListUsers(c *gin.Context) {
	q := middleware.GetBind[dto.AdminListUsersReq](c)
	global.Log.Info(q)
	resp, err := uh.userLogic.AdminListUsers(c.Request.Context(), q)
	response.Response(c, resp, err)
}

func (uh *UserHandler) AdminPromoteToAdmin(c *gin.Context) {
	uri := middleware.GetBind[dto.AdminPromoteUri](c)
	global.Log.Info(uri)
	msg, err := uh.userLogic.AdminPromoteToAdmin(c.Request.Context(), uri.UserID)
	if err != nil {
		response.Response(c, nil, err)
		return
	}
	response.Response(c, gin.H{"message": msg}, nil)
}
