package handler

import (
	"context"
	"net/http"
	"nurture/internal/global"
	"nurture/internal/logic"
	"nurture/internal/pkg/jwtx"
	"nurture/internal/pkg/wsx"

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

func (h *WebSocketHandler) Connect(c *gin.Context) {
	tokenStr := c.Query("token")
	partnerID := c.Query("user_id")

	if tokenStr == "" || partnerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token or user_id"})
		return
	}

	// Verify token
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.UserID

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &wsx.Client{
		Hub:       global.WS,
		Conn:      conn,
		Send:      make(chan []byte, 256),
		UserID:    userID,
		PartnerID: partnerID,
	}

	client.Hub.RegisterClient(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *WebSocketHandler) ConnectGroups(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing token"})
		return
	}
	claims, err := jwtx.ParseTokenString(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID := claims.UserID
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	chatLogic := logic.NewChatLogic()
	client := &wsx.GroupClient{
		Hub:    global.GWS,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		UserID: userID,
		OnSubscribe: func(ctx context.Context, groupID string) error {
			return chatLogic.CheckMember(ctx, userID, groupID)
		},
		OnSend: func(ctx context.Context, groupID, messageID, msgType, content string, now int64) error {
			return chatLogic.SaveMessage(ctx, userID, groupID, messageID, msgType, content, now)
		},
	}
	client.Hub.RegisterClient(client)
	go client.WritePump()
	go client.ReadPump()
}
