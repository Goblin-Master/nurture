package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (uh *UserHandler) UpdateProfile(c *gin.Context) {
	cr := middleware.GetBind[dto.UpdateUserAdditionReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.UpdateProfile(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) UpdateAvatar(c *gin.Context) {
	cr := middleware.GetBind[dto.UpdateAvatarReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.UpdateAvatar(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) MyProfile(c *gin.Context) {
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.MyProfile(c.Request.Context(), userID)
	response.Response(c, resp, err)
}
