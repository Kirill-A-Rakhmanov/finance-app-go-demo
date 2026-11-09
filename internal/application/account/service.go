package account

import (
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"
)

type Service struct {
	txManager       transaction.Manager
	accountRepo     repository.AccountRepository
	transactionRepo repository.TransactionRepository
	userRepo        repository.UserRepository
}

func NewService(
	txManager transaction.Manager,
	accountRepo repository.AccountRepository,
	transactionRepo repository.TransactionRepository,
	userRepo repository.UserRepository,
) *Service {
	return &Service{
		txManager:       txManager,
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
	}
}
