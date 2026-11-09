package user

import (
	"finance-app-go-demo/internal/application/user"
)

type Handler struct {
	userService *user.Service
}

func NewHandler(userService *user.Service) *Handler {
	return &Handler{
		userService: userService,
	}
}
