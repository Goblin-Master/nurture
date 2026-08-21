package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"nurture/internal/chat/constant"
	"nurture/internal/chat/logic"
	"nurture/internal/chat/session"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/response"
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

func (h *ChatHandler) ConnectDirect(c *gin.Context) {
	tokenStr := c.Query("token")
	partnerID := c.Query("user_id")
	if tokenStr == "" || partnerID == "" {
		response.Response(c, nil, ErrTokenEmpty)
		return
	}
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		response.Response(c, nil, ErrTokenInvalid)
		return
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	userID := claims.UserID
	client := session.NewClient(h.hub, conn, userID, constant.ChannelDirect)
	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.handleDirectMessage(ctx, client, partnerID, message)
	})
}

func (h *ChatHandler) ConnectGroup(c *gin.Context) {
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
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	userID := claims.UserID
	client := session.NewClient(h.hub, conn, userID, constant.ChannelGroup)
	client.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump(func(message []byte) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		h.handleGroupMessage(ctx, client, message)
	})
}

func (h *ChatHandler) handleDirectMessage(ctx context.Context, client *session.Client, partnerID string, message []byte) {
	result, _ := h.chatLogic.HandleDirectMessage(ctx, client.UserID, partnerID, message)
	if result.Ack != nil {
		writeDirectAck(client, *result.Ack)
	}
}

func (h *ChatHandler) handleGroupMessage(ctx context.Context, client *session.Client, message []byte) {
	result := h.chatLogic.HandleGroupMessage(ctx, client.UserID, message)
	for _, groupID := range result.Subscribe {
		client.Hub.Subscribe(client, groupID)
	}
	for _, groupID := range result.Unsubscribe {
		client.Hub.Unsubscribe(client, groupID)
	}
	if result.Ack != nil {
		writeGroupAck(client, *result.Ack)
	}
}

func writeDirectAck(client *session.Client, ack logic.DirectAckMessage) {
	b, err := json.Marshal(ack)
	if err != nil {
		return
	}
	client.TrySend(b)
}

func writeGroupAck(client *session.Client, ack logic.GroupAckMessage) {
	b, err := json.Marshal(ack)
	if err != nil {
		return
	}
	client.TrySend(b)
}
