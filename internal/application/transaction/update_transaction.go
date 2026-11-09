package transaction

import (
	"context"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/validation"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type UpdateTransactionInput struct {
	TransactionID           uuid.UUID
	UserID                  uuid.UUID
	AccountID               uuid.UUID
	TransactionCategoryName *string
	CounterpartyName        *string
	Amount                  decimal.Decimal
	Name                    *string
	Date                    time.Time
	Type                    enum.TransactionType
	IsConfirmed             bool
}

type UpdateTransactionResult struct {
	Transaction *model.Transaction
}

func (s *Service) UpdateTransaction(ctx context.Context, input UpdateTransactionInput) (*UpdateTransactionResult, error) {
	err := input.Validate()
	if err != nil {
		return nil, err
	}

	var result *UpdateTransactionResult

	err = s.txManager.Wrap(ctx, func(ctx context.Context) error {
		targetAccount, err := s.accountRepo.GetByIDAndUserID(ctx, input.AccountID, input.UserID)
		if err != nil {
			return err
		}
		if targetAccount == nil {
			return apperror.ErrAccountNotFound
		}

		transaction, err := s.transactionRepo.GetByIDAndUserID(ctx, input.TransactionID, input.UserID)
		if err != nil {
			return err
		}
		if transaction == nil {
			return ErrTransactionNotFound
		}
		if transaction.TransferID != nil {
			return ErrTransferUpdateNotAllowed
		}

		sourceAccount := targetAccount
		if transaction.AccountID != targetAccount.ID {
			sourceAccount, err = s.accountRepo.GetByIDAndUserID(ctx, transaction.AccountID, input.UserID)
			if err != nil {
				return err
			}
			if sourceAccount == nil {
				return apperror.ErrAccountNotFound
			}
			if sourceAccount.Currency.Code != targetAccount.Currency.Code {
				return ErrMoveToOtherCurrencyNotAllowed
			}
		}

		oldEffect := transaction.BalanceEffect()
		newEffect := model.BalanceEffect(input.Type, input.Amount, input.IsConfirmed)

		if err := s.ensureBalancesStayNonNegative(ctx, sourceAccount, targetAccount, oldEffect, newEffect); err != nil {
			return err
		}

		categoryID, err := s.resolveTransactionCategoryID(ctx, input.UserID, input.Type, input.TransactionCategoryName)
		if err != nil {
			return err
		}

		counterpartyID, err := s.resolveCounterpartyID(ctx, input.UserID, input.CounterpartyName)
		if err != nil {
			return err
		}

		transaction.AccountID = targetAccount.ID
		transaction.Amount = input.Amount
		transaction.Name = input.Name
		transaction.CounterpartyID = counterpartyID
		transaction.Date = input.Date
		transaction.Type = input.Type
		transaction.TransactionCategoryID = categoryID
		transaction.IsConfirmed = input.IsConfirmed

		transaction, err = s.transactionRepo.Update(ctx, transaction)
		if err != nil {
			return err
		}

		result = &UpdateTransactionResult{Transaction: transaction}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *Service) ensureBalancesStayNonNegative(
	ctx context.Context,
	source *model.Account,
	target *model.Account,
	oldEffect decimal.Decimal,
	newEffect decimal.Decimal,
) error {
	changes := map[uuid.UUID]decimal.Decimal{source.ID: oldEffect.Neg()}
	changes[target.ID] = changes[target.ID].Add(newEffect)

	checked := make([]*model.Account, 0, 2)
	for _, account := range []*model.Account{source, target} {
		if !account.IsNegativeBalanceEnabled && changes[account.ID].IsNegative() && !slices.Contains(checked, account) {
			checked = append(checked, account)
		}
	}

	slices.SortFunc(checked, func(a, b *model.Account) int {
		return strings.Compare(model.AccountBalanceLockKey(a.ID), model.AccountBalanceLockKey(b.ID))
	})

	for _, account := range checked {
		if err := s.txManager.Lock(ctx, model.AccountBalanceLockKey(account.ID)); err != nil {
			return err
		}
	}

	for _, account := range checked {
		balance, err := s.transactionRepo.GetBalanceByAccountID(ctx, account.ID)
		if err != nil {
			return err
		}

		if balance.Add(changes[account.ID]).IsNegative() {
			return apperror.ErrNegativeBalanceNotAllowed
		}
	}

	return nil
}

func (i *UpdateTransactionInput) Validate() error {
	v := apperror.NewValidationError()
	v.AddIf(i.AccountID == uuid.Nil, "accountId", "is required")
	validation.Amount(v, "amount", i.Amount)
	v.MaxLengthPtr("name", i.Name, model.TransactionNameMaxLen)
	v.MaxLengthPtr("transactionCategoryName", i.TransactionCategoryName, model.TransactionCategoryNameMaxLen)
	v.MaxLengthPtr("counterpartyName", i.CounterpartyName, model.CounterpartyNameMaxLen)
	return v.ErrorOrNil()
}
