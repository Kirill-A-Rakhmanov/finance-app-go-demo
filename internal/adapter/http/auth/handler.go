package auth

import (
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/application/auth"
)

type Handler struct {
	authService *auth.Service
	cookies     cookie.Config
}

func NewHandler(authService *auth.Service, cookies cookie.Config) *Handler {
	return &Handler{
		authService: authService,
		cookies:     cookies,
	}
}
