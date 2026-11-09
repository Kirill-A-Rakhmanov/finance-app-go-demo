package transaction

import (
	"context"
	"strings"

	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
)

func normalizeName(name *string) *string {
	if name == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*name)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func (s *Service) resolveTransactionCategoryID(
	ctx context.Context,
	userID uuid.UUID,
	transactionType enum.TransactionType,
	name *string,
) (*uuid.UUID, error) {
	name = normalizeName(name)
	if name == nil {
		return nil, nil
	}

	categoryType := enum.Expense
	if transactionType == enum.TransactionIncome {
		categoryType = enum.Income
	}

	category, err := s.transactionCategoryRepo.GetByUserIDAndTypeAndName(ctx, userID, categoryType, *name)
	if err != nil {
		return nil, err
	}
	if category != nil {
		return &category.ID, nil
	}

	category, err = s.transactionCategoryRepo.Create(ctx, model.NewTransactionCategory(
		userID,
		nil,
		*name,
		model.DefaultTransactionCategoryIcon,
		model.DefaultTransactionCategoryColor,
		categoryType,
	))
	if err != nil {
		return nil, err
	}

	return &category.ID, nil
}

func (s *Service) resolveCounterpartyID(ctx context.Context, userID uuid.UUID, name *string) (*uuid.UUID, error) {
	name = normalizeName(name)
	if name == nil {
		return nil, nil
	}

	counterparty, err := s.counterpartyRepo.GetByUserIDAndName(ctx, userID, *name)
	if err != nil {
		return nil, err
	}
	if counterparty != nil {
		return &counterparty.ID, nil
	}

	counterparty, err = s.counterpartyRepo.Create(ctx, model.NewCounterparty(userID, *name))
	if err != nil {
		return nil, err
	}

	return &counterparty.ID, nil
}
