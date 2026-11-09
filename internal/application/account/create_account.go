package account

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateAccountInput struct {
	UserID                   uuid.UUID
	Name                     string
	Icon                     string
	Currency                 string
	IsNegativeBalanceEnabled bool
	IsHidden                 bool
	IsExcludedFromCount      bool
	InitialBalance           decimal.Decimal
}

type CreateAccountResult struct {
	Account *model.Account
}

func (s *Service) CreateAccount(ctx context.Context, input CreateAccountInput) (*CreateAccountResult, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	exist, err := s.userRepo.ExistsByID(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if !exist {
		return nil, apperror.ErrUserNotFound
	}

	currency, err := valueObject.NewCurrency(valueObject.CurrencyCode(input.Currency))
	if err != nil {
		return nil, apperror.Invalid(err.Error())
	}

	account := model.NewAccount(
		input.UserID,
		input.Name,
		input.Icon,
		currency,
		input.IsNegativeBalanceEnabled,
		input.IsHidden,
		input.IsExcludedFromCount,
		input.InitialBalance,
	)

	account, err = s.accountRepo.Create(ctx, account)
	if err != nil {
		return nil, err
	}

	return &CreateAccountResult{
		Account: account,
	}, nil
}

func (i *CreateAccountInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.UserID == uuid.Nil, "userId", "is required")
	v.AddIf(i.Name == "", "name", "is required")
	v.MaxLength("name", i.Name, model.AccountNameMaxLen)
	v.MaxLength("icon", i.Icon, model.AccountIconMaxLen)
	return v.ErrorOrNil()
}
