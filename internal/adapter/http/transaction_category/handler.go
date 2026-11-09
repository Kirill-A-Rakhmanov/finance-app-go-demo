package transaction_category

import "finance-app-go-demo/internal/application/transaction_category"

type Handler struct {
	transactionCategoryService *transaction_category.Service
}

func NewHandler(transactionCategoryService *transaction_category.Service) *Handler {
	return &Handler{
		transactionCategoryService: transactionCategoryService,
	}
}
