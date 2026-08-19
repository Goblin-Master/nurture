package handler

import (
	"nurture/internal/chat/logic"
	"nurture/internal/chat/session"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	chatLogic logic.IChatLogic
	hub       *session.Hub
}

func NewChatHandler(chatLogic logic.IChatLogic, hub *session.Hub) *ChatHandler {
	return &ChatHandler{
		chatLogic: chatLogic,
		hub:       hub,
	}
}

func (h *ChatHandler) bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return false
	}
	return true
}

func (h *ChatHandler) bindQuery(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return false
	}
	return true
}

func (h *ChatHandler) bindURI(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindUri(dst); err != nil {
		response.Response(c, nil, err)
		c.Abort()
		return false
	}
	return true
}
