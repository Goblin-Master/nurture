package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/session"
	"time"

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

func (h *ChatHandler) ConnectDirect(c *gin.Context) {
	tokenStr := c.Query("token")
	partnerID := c.Query("user_id")
	if tokenStr == "" || partnerID == "" {
		h.respond(c, nil, ErrTokenEmpty)
		return
	}
	userID, err := h.parseToken(tokenStr)
	if err != nil {
		h.respond(c, nil, ErrTokenInvalid)
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := session.NewClient(h.hub, conn, userID, constant.ChannelDirect)
	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		message = bytes.TrimSpace(bytes.ReplaceAll(message, newline, space))
		client.Hub.SendToUser(constant.ChannelDirect, partnerID, message)
	})
}

func (h *ChatHandler) ConnectGroup(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		h.respond(c, nil, ErrTokenEmpty)
		return
	}
	userID, err := h.parseToken(tokenStr)
	if err != nil {
		h.respond(c, nil, ErrTokenInvalid)
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	client := session.NewClient(h.hub, conn, userID, constant.ChannelGroup)
	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.handleGroupMessage(ctx, client, message)
	})
}

func (h *ChatHandler) handleGroupMessage(ctx context.Context, client *session.Client, message []byte) {
	result := h.chatLogic.HandleGroupMessage(ctx, client.UserID, message)
	for _, groupID := range result.Subscribe {
		client.Hub.Subscribe(client, groupID)
	}
	for _, groupID := range result.Unsubscribe {
		client.Hub.Unsubscribe(client, groupID)
	}
	if len(result.Broadcast) > 0 && result.BroadcastGroupID != "" {
		client.Hub.Broadcast(result.BroadcastGroupID, result.Broadcast)
	}
	if result.Ack != nil {
		writeGroupAck(client, *result.Ack)
	}
}

func writeGroupAck(client *session.Client, ack logic.GroupAckMessage) {
	b, err := json.Marshal(ack)
	if err != nil {
		return
	}
	client.TrySend(b)
}
