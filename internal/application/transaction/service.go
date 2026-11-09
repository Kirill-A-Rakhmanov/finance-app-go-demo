package transaction

import (
	"finance-app-go-demo/internal/application/query"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"
)

type Service struct {
	transactionRepo         repository.TransactionRepository
	accountRepo             repository.AccountRepository
	transactionCategoryRepo repository.TransactionCategoryRepository
	counterpartyRepo        repository.CounterpartyRepository
	txManager               transaction.Manager
	transactionQueryService query.TransactionQueryService
}

func NewService(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	transactionCategoryRepo repository.TransactionCategoryRepository,
	counterpartyRepo repository.CounterpartyRepository,
	txManager transaction.Manager,
	transactionQueryService query.TransactionQueryService,
) *Service {
	return &Service{
		transactionRepo:         transactionRepo,
		accountRepo:             accountRepo,
		transactionCategoryRepo: transactionCategoryRepo,
		counterpartyRepo:        counterpartyRepo,
		txManager:               txManager,
		transactionQueryService: transactionQueryService,
	}
}
