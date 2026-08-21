package handler

import (
	"nurture/internal/chat/dto"
	"nurture/internal/middleware"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *ChatHandler) ListDirectMessages(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatDirectMessageUserUri](c)
	q := middleware.GetBind[dto.ChatDirectMessageListReq](c)
	userID := jwtx.GetUserID(c)
	resp, err := h.chatLogic.ListDirectMessages(c.Request.Context(), userID, uri, q)
	response.Response(c, resp, err)
}

func (h *ChatHandler) MarkDirectSeen(c *gin.Context) {
	uri := middleware.GetBind[dto.ChatDirectMessageUserUri](c)
	userID := jwtx.GetUserID(c)
	err := h.chatLogic.MarkDirectSeen(c.Request.Context(), userID, uri.UserID, 0)
	response.Response(c, nil, err)
}
