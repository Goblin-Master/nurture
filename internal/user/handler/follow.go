package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (uh *UserHandler) Follow(c *gin.Context) {
	uri := middleware.GetBind[dto.FollowReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.Follow(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (uh *UserHandler) Unfollow(c *gin.Context) {
	uri := middleware.GetBind[dto.FollowReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.Unfollow(c.Request.Context(), userID, uri)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ListFollowing(c *gin.Context) {
	q := middleware.GetBind[dto.FollowingListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.ListFollowing(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}

func (uh *UserHandler) ListFollowers(c *gin.Context) {
	q := middleware.GetBind[dto.FollowersListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.ListFollowers(c.Request.Context(), userID, q)
	response.Response(c, resp, err)
}
