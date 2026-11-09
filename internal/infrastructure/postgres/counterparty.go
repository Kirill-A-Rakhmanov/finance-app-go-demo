package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/counterparty"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"

	"github.com/google/uuid"
)

func toCounterpartyEntity(tc *model.Counterparty) *ent.Counterparty {
	if tc == nil {
		return nil
	}

	return &ent.Counterparty{
		ID:        tc.ID,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
		UserID:    tc.UserID,
		Name:      tc.Name,
	}
}

func toCounterpartyModel(tc *ent.Counterparty) *model.Counterparty {
	if tc == nil {
		return nil
	}

	return &model.Counterparty{
		ID:        tc.ID,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
		UserID:    tc.UserID,
		Name:      tc.Name,
	}
}

type counterpartyRepository struct {
	client *ent.Client
}

func NewCounterpartyRepository(clent *ent.Client) repository.CounterpartyRepository {
	return &counterpartyRepository{client: clent}
}

func (r *counterpartyRepository) Create(ctx context.Context, tc *model.Counterparty) (*model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.Counterparty.Create().
		SetUserID(tc.UserID).
		SetName(tc.Name)

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toCounterpartyModel(entity), nil
}

func (r *counterpartyRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.Counterparty.Delete().
		Where(counterparty.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *counterpartyRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.Counterparty.Query().
		Where(counterparty.IDEQ(ID)).
		Exist(ctx)
}

func (r *counterpartyRepository) ExistsByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.Counterparty.Query().
		Where(
			counterparty.UserIDEQ(userID),
			counterparty.Name(name),
		).
		Exist(ctx)
}

func (r *counterpartyRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Counterparty.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toCounterpartyModel(entity), nil
}

func (r *counterpartyRepository) GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Counterparty.Query().
		Where(
			counterparty.IDEQ(ID),
			counterparty.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toCounterpartyModel(entity), nil
}

func (r *counterpartyRepository) GetByUserIDAndName(ctx context.Context, userID uuid.UUID, name string) (*model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.Counterparty.Query().
		Where(
			counterparty.UserIDEQ(userID),
			counterparty.Name(name),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toCounterpartyModel(entity), nil
}

func (r *counterpartyRepository) GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Counterparty.Query().
		Where(
			counterparty.IDIn(IDs...),
			counterparty.UserIDEQ(userID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Counterparty, len(entities))

	for i, e := range entities {
		result[i] = *toCounterpartyModel(e)
	}

	return result, nil
}

func (r *counterpartyRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.Counterparty.Query().
		Where(counterparty.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.Counterparty, len(entities))

	for i, e := range entities {
		result[i] = *toCounterpartyModel(e)
	}

	return result, nil
}

func (r *counterpartyRepository) Update(ctx context.Context, tc *model.Counterparty) (*model.Counterparty, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.Counterparty.UpdateOneID(tc.ID).
		SetUserID(tc.UserID).
		SetName(tc.Name)

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toCounterpartyModel(entity), nil
}
