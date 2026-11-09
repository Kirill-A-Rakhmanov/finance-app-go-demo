package transfer

import "finance-app-go-demo/internal/application/transfer"

type Handler struct {
	transferService *transfer.Service
}

func NewHandler(transferService *transfer.Service) *Handler {
	return &Handler{
		transferService: transferService,
	}
}
