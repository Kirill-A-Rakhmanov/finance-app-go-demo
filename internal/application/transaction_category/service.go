package transaction_category

import "finance-app-go-demo/internal/domain/repository"

type Service struct {
	transactionCategoryRepo repository.TransactionCategoryRepository
}

func New(transactionCategoryRepo repository.TransactionCategoryRepository) *Service {
	return &Service{
		transactionCategoryRepo: transactionCategoryRepo,
	}
}
