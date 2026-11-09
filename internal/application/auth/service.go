package auth

import (
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"
)

type Service struct {
	userRepo                repository.UserRepository
	transactionCategoryRepo repository.TransactionCategoryRepository
	featureFlagRepo         repository.FeatureFlagRepository
	tokenService            TokenService
	passwordService         PasswordService
	txManager               transaction.Manager
	sessionStorage          SessionStorage
	recoveryStorage         RecoveryStorage
	recoveryNotifier        RecoveryNotifier
	registrationStorage     RegistrationStorage
	registrationNotifier    RegistrationNotifier
	failureLimiter          FailureLimiter
}

func NewService(
	userRepo repository.UserRepository,
	transactionCategoryRepo repository.TransactionCategoryRepository,
	featureFlagRepo repository.FeatureFlagRepository,
	tokenService TokenService,
	passwordService PasswordService,
	txManager transaction.Manager,
	sessionStorage SessionStorage,
	recoveryStorage RecoveryStorage,
	recoveryNotifier RecoveryNotifier,
	registrationStorage RegistrationStorage,
	registrationNotifier RegistrationNotifier,
	failureLimiter FailureLimiter,
) *Service {
	return &Service{
		userRepo:                userRepo,
		transactionCategoryRepo: transactionCategoryRepo,
		featureFlagRepo:         featureFlagRepo,
		tokenService:            tokenService,
		passwordService:         passwordService,
		txManager:               txManager,
		sessionStorage:          sessionStorage,
		recoveryStorage:         recoveryStorage,
		recoveryNotifier:        recoveryNotifier,
		registrationStorage:     registrationStorage,
		registrationNotifier:    registrationNotifier,
		failureLimiter:          failureLimiter,
	}
}
