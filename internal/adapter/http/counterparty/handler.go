package counterparty

import "finance-app-go-demo/internal/application/counterparty"

type Handler struct {
	counterpartyService *counterparty.Service
}

func NewHandler(counterpartyService *counterparty.Service) *Handler {
	return &Handler{
		counterpartyService: counterpartyService,
	}
}
