package handler

import (
	"nurture/internal/chat/logic"
	"nurture/internal/chat/session"
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
