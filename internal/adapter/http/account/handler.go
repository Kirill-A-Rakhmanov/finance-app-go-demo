package account

import "finance-app-go-demo/internal/application/account"

type Handler struct {
	accountService *account.Service
}

func NewHandler(accountService *account.Service) *Handler {
	return &Handler{
		accountService: accountService,
	}
}
