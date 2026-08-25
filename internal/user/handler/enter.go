package handler

import (
	"nurture/internal/user/logic"
)

type UserHandler struct {
	userLogic logic.IUserLogic
}

func NewUserHandler(userLogic logic.IUserLogic) *UserHandler {
	return &UserHandler{
		userLogic: userLogic,
	}
}
