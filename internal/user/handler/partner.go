package handler

import (
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
	"nurture/internal/user/dto"

	"github.com/gin-gonic/gin"
)

func (uh *UserHandler) BindPartner(c *gin.Context) {
	cr := middleware.GetBind[dto.PartnerBindReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.BindPartner(c.Request.Context(), userID, cr)
	response.Response(c, resp, err)
}

func (uh *UserHandler) GetPartner(c *gin.Context) {
	userID := jwtx.GetUserID(c)
	resp, err := uh.userLogic.GetPartner(c.Request.Context(), userID)
	response.Response(c, resp, err)
}
