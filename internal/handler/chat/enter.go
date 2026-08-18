package chat

import chatlogic "nurture/internal/logic/chat"

type Handler struct {
	chatLogic *chatlogic.Logic
	hub       *sessionHub
}

func NewHandler() *Handler {
	hub := newSessionHub()
	go hub.run()
	return &Handler{
		chatLogic: chatlogic.NewLogic(),
		hub:       hub,
	}
}
