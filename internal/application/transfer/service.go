package transfer

import (
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"
)

type Service struct {
	transferRepo    repository.TransferRepository
	transactionRepo repository.TransactionRepository
	accountRepo     repository.AccountRepository
	txManager       transaction.Manager
}

func NewService(
	transferRepo repository.TransferRepository,
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	txManager transaction.Manager,
) *Service {
	return &Service{
		transferRepo:    transferRepo,
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		txManager:       txManager,
	}
}
