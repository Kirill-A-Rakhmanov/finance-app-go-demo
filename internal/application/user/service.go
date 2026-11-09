package user

import (
	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/internal/domain/repository"
)

type Service struct {
	userRepo        repository.UserRepository
	transactionRepo repository.TransactionRepository
	passwordService auth.PasswordService
	sessionRevoker  SessionRevoker
	failureLimiter  auth.FailureLimiter
}

func NewService(
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	passwordService auth.PasswordService,
	sessionRevoker SessionRevoker,
	failureLimiter auth.FailureLimiter,
) *Service {
	return &Service{
		userRepo:        userRepo,
		transactionRepo: transactionRepo,
		passwordService: passwordService,
		sessionRevoker:  sessionRevoker,
		failureLimiter:  failureLimiter,
	}
}
