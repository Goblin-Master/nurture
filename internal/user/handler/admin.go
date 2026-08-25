package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

// admin
func (uh *UserHandler) AdminListUsers(c *gin.Context) {
	q := middleware.GetBind[dto.AdminListUsersReq](c)
	resp, err := uh.userLogic.AdminListUsers(c.Request.Context(), q)
	response.Response(c, resp, err)
}

func (uh *UserHandler) AdminPromoteToAdmin(c *gin.Context) {
	uri := middleware.GetBind[dto.AdminPromoteUri](c)
	msg, err := uh.userLogic.AdminPromoteToAdmin(c.Request.Context(), uri.UserID)
	if err != nil {
		response.Response(c, nil, err)
		return
	}
	response.Response(c, gin.H{"message": msg}, nil)
}
