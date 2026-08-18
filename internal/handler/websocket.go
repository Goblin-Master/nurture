package handler

import (
	"bytes"
	"net/http"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/realtimex"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct{}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func (h *WebSocketHandler) ConnectDirect(c *gin.Context) {
	tokenStr := c.Query("token")
	partnerID := c.Query("user_id")

	if tokenStr == "" || partnerID == "" {
		response.Response(c, nil, ErrTokenEmpty)
		return
	}

	// Verify token
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		response.Response(c, nil, ErrTokenInvalid)
		return
	}

	userID := claims.UserID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := realtimex.NewClient(global.Realtime, conn, userID, realtimex.ChannelDirect)
	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		client.Hub.SendToUser(realtimex.ChannelDirect, partnerID, message)
	})
}

func (h *WebSocketHandler) ConnectGroup(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		response.Response(c, nil, ErrTokenEmpty)
		return
	}
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		response.Response(c, nil, ErrTokenInvalid)
		return
	}
	userID := claims.UserID
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	chatLogic := logic.NewChatLogic()
	client := realtimex.NewClient(global.Realtime, conn, userID, realtimex.ChannelGroup)
	client.Hub.Register(client)
	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		h.handleGroupMessage(client, chatLogic, message)
	})
}
