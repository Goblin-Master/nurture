package handler

import (
	"nurture/internal/chat/logic"
	"nurture/internal/chat/session"

	"github.com/gin-gonic/gin"
)

type RespondFunc func(c *gin.Context, resp interface{}, err error)

type GetUserIDFunc func(c *gin.Context) string

type ParseTokenFunc func(token string) (string, error)

type ChatHandler struct {
	chatLogic  logic.IChatLogic
	hub        *session.Hub
	getUserID  GetUserIDFunc
	parseToken ParseTokenFunc
	respond    RespondFunc
}

func NewChatHandler(chatLogic logic.IChatLogic, hub *session.Hub, getUserID GetUserIDFunc, parseToken ParseTokenFunc, respond RespondFunc) *ChatHandler {
	return &ChatHandler{
		chatLogic:  chatLogic,
		hub:        hub,
		getUserID:  getUserID,
		parseToken: parseToken,
		respond:    respond,
	}
}

func (h *ChatHandler) bindJSON(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		h.respond(c, nil, err)
		c.Abort()
		return false
	}
	return true
}

func (h *ChatHandler) bindQuery(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindQuery(dst); err != nil {
		h.respond(c, nil, err)
		c.Abort()
		return false
	}
	return true
}

func (h *ChatHandler) bindURI(c *gin.Context, dst interface{}) bool {
	if err := c.ShouldBindUri(dst); err != nil {
		h.respond(c, nil, err)
		c.Abort()
		return false
	}
	return true
}
