package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/account"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/pkg/transaction"
	"fmt"

	"github.com/google/uuid"
)

// mapper
func toAccountEntity(a *model.Account) *ent.Account {
	if a == nil {
		return nil
	}

	return &ent.Account{
		ID:                       a.ID,
		CreatedAt:                a.CreatedAt,
		UpdatedAt:                a.UpdatedAt,
		UserID:                   a.UserID,
		Name:                     a.Name,
		Icon:                     a.Icon,
		Currency:                 string(a.Currency.Code),
		IsNegativeBalanceEnabled: a.IsNegativeBalanceEnabled,
		IsHidden:                 a.IsHidden,
		IsExcludedFromCount:      a.IsExcludedFromCount,
		Adjustment:               a.Adjustment,
	}
}

func toAccountModel(a *ent.Account) (*model.Account, error) {
	if a == nil {
		return nil, nil
	}

	currecy, err := valueObject.NewCurrency(valueObject.CurrencyCode(a.Currency))
	if err != nil {
		return nil, fmt.Errorf("failed to map account %s: %w", a.ID, err)
	}

	return &model.Account{
		ID:                       a.ID,
		CreatedAt:                a.CreatedAt,
		UpdatedAt:                a.UpdatedAt,
		UserID:                   a.UserID,
		Name:                     a.Name,
		Icon:                     a.Icon,
		Currency:                 currecy,
		IsNegativeBalanceEnabled: a.IsNegativeBalanceEnabled,
		IsHidden:                 a.IsHidden,
		IsExcludedFromCount:      a.IsExcludedFromCount,
		Adjustment:               a.Adjustment,
	}, nil
}

func mapAccounts(entities []*ent.Account) ([]model.Account, error) {
	result := make([]model.Account, len(entities))

	for i, e := range entities {
		m, err := toAccountModel(e)
		if err != nil {
			return nil, err
		}

		result[i] = *m
	}

	return result, nil
}

type accountRepository struct {
	client *ent.Client
}

func NewAccountRepository(client *ent.Client) repository.AccountRepository {
	return &accountRepository{client: client}
}

func (r *accountRepository) Create(ctx context.Context, a *model.Account) (*model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.Account.Create().
		SetUserID(a.UserID).
		SetName(a.Name).
		SetIcon(a.Icon).
		SetCurrency(string(a.Currency.Code)).
		SetIsNegativeBalanceEnabled(a.IsNegativeBalanceEnabled).
		SetIsHidden(a.IsHidden).
		SetIsExcludedFromCount(a.IsExcludedFromCount).
		SetAdjustment(a.Adjustment)

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toAccountModel(entity)
}

func (r *accountRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.Account.Delete().
		Where(account.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *accountRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.Account.Query().
		Where(account.IDEQ(ID)).
		Exist(ctx)
}

func (r *accountRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Account.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toAccountModel(entity)
}

func (r *accountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Account.Query().
		Where(account.UserID(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return mapAccounts(entities)
}

func (r *accountRepository) GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Account.Query().
		Where(
			account.IDEQ(ID),
			account.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toAccountModel(entity)
}

func (r *accountRepository) GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Account.Query().
		Where(
			account.IDIn(IDs...),
			account.UserIDEQ(userID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	return mapAccounts(entities)
}

func (r *accountRepository) Update(ctx context.Context, a *model.Account) (*model.Account, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.Account.UpdateOneID(a.ID).
		SetUserID(a.UserID).
		SetName(a.Name).
		SetIcon(a.Icon).
		SetCurrency(string(a.Currency.Code)).
		SetIsNegativeBalanceEnabled(a.IsNegativeBalanceEnabled).
		SetIsHidden(a.IsHidden).
		SetIsExcludedFromCount(a.IsExcludedFromCount).
		SetAdjustment(a.Adjustment)

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toAccountModel(entity)
}
