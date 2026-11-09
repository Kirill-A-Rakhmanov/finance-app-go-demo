package transaction

import "finance-app-go-demo/internal/application/transaction"

type Handler struct {
	transactionService *transaction.Service
}

func NewHandler(transactionService *transaction.Service) *Handler {
	return &Handler{
		transactionService: transactionService,
	}
}
