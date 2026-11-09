package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/transfer"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"

	"github.com/google/uuid"
)

func toTransferEntity(t *model.Transfer) *ent.Transfer {
	if t == nil {
		return nil
	}

	return &ent.Transfer{
		ID:            t.ID,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		FromAmount:    t.FromAmount,
		ToAmount:      t.ToAmount,
		Date:          t.Date,
	}
}

func toTransferModel(t *ent.Transfer) *model.Transfer {
	if t == nil {
		return nil
	}

	return &model.Transfer{
		ID:            t.ID,
		CreatedAt:     t.CreatedAt,
		UpdatedAt:     t.UpdatedAt,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		FromAmount:    t.FromAmount,
		ToAmount:      t.ToAmount,
		Date:          t.Date,
	}
}

type transferRepository struct {
	client *ent.Client
}

func NewTransferRepository(client *ent.Client) repository.TransferRepository {
	return &transferRepository{client: client}
}

func (r *transferRepository) Create(ctx context.Context, t *model.Transfer) (*model.Transfer, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.Transfer.Create().
		SetFromAccountID(t.FromAccountID).
		SetToAccountID(t.ToAccountID).
		SetFromAmount(t.FromAmount).
		SetToAmount(t.ToAmount).
		SetDate(t.Date)

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransferModel(entity), nil
}

func (r *transferRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.Transfer.Delete().
		Where(transfer.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *transferRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.Transfer.Query().
		Where(transfer.IDEQ(ID)).
		Exist(ctx)
}

func (r *transferRepository) GetByFromAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transfer, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Transfer.Query().
		Where(transfer.FromAccountIDEQ(accountID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transfer, len(entities))

	for i, e := range entities {
		result[i] = *toTransferModel(e)
	}

	return result, nil
}

func (r *transferRepository) GetByToAccountID(ctx context.Context, accountID uuid.UUID) ([]model.Transfer, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Transfer.Query().
		Where(transfer.ToAccountIDEQ(accountID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Transfer, len(entities))

	for i, e := range entities {
		result[i] = *toTransferModel(e)
	}

	return result, nil
}

func (r *transferRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.Transfer, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Transfer.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransferModel(entity), nil
}

func (r *transferRepository) Update(ctx context.Context, t *model.Transfer) (*model.Transfer, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.Transfer.UpdateOneID(t.ID).
		SetFromAccountID(t.FromAccountID).
		SetToAccountID(t.ToAccountID).
		SetFromAmount(t.FromAmount).
		SetToAmount(t.ToAmount).
		SetDate(t.Date)

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransferModel(entity), nil
}
