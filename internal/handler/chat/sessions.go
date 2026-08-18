package chat

import (
	"bytes"
	"net/http"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

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

func (h *Handler) OpenDirectSession(c *gin.Context) {
	tokenStr := c.Query("token")
	partnerID := c.Query("user_id")

	if tokenStr == "" || partnerID == "" {
		response.Response(c, nil, errTokenEmpty)
		return
	}

	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		response.Response(c, nil, errTokenInvalid)
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := newSession(h.hub, conn, claims.UserID, directChannel)
	client.hub.registerSession(client)

	go client.writeMessages()
	go client.readMessages(func(message []byte) {
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		client.hub.sendToUserChannel(directChannel, partnerID, message)
	})
}

func (h *Handler) OpenGroupSession(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		response.Response(c, nil, errTokenEmpty)
		return
	}
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		response.Response(c, nil, errTokenInvalid)
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := newSession(h.hub, conn, claims.UserID, groupChannel)
	client.hub.registerSession(client)
	go client.writeMessages()
	go client.readMessages(func(message []byte) {
		h.handleGroupMessage(client, message)
	})
}
