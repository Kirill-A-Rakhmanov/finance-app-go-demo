package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/transactioncategory"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/pkg/transaction"

	"github.com/google/uuid"
)

func toTransactionCategoryEntity(tc *model.TransactionCategory) *ent.TransactionCategory {
	if tc == nil {
		return nil
	}

	return &ent.TransactionCategory{
		ID:        tc.ID,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
		UserID:    tc.UserID,
		ParentID:  tc.ParentID,
		Name:      tc.Name,
		Slug:      tc.Slug,
		Icon:      tc.Icon,
		Color:     tc.Color,
		IsDefault: tc.IsDefault,
		Type:      transactioncategory.Type(tc.Type),
	}
}

func toTransactionCategoryModel(tc *ent.TransactionCategory) *model.TransactionCategory {
	if tc == nil {
		return nil
	}

	return &model.TransactionCategory{
		ID:        tc.ID,
		CreatedAt: tc.CreatedAt,
		UpdatedAt: tc.UpdatedAt,
		UserID:    tc.UserID,
		ParentID:  tc.ParentID,
		Name:      tc.Name,
		Slug:      tc.Slug,
		Icon:      tc.Icon,
		Color:     tc.Color,
		IsDefault: tc.IsDefault,
		Type:      enum.TransactionCategoryType(tc.Type),
	}
}

type transactionCategoryRepository struct {
	client *ent.Client
}

func NewTransactionCategoryRepository(clent *ent.Client) repository.TransactionCategoryRepository {
	return &transactionCategoryRepository{client: clent}
}

func (r *transactionCategoryRepository) Create(ctx context.Context, tc *model.TransactionCategory) (*model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.TransactionCategory.Create().
		SetUserID(tc.UserID).
		SetNillableParentID(tc.ParentID).
		SetName(tc.Name).
		SetNillableSlug(tc.Slug).
		SetIcon(tc.Icon).
		SetColor(tc.Color).
		SetIsDefault(tc.IsDefault).
		SetType(transactioncategory.Type(tc.Type))

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransactionCategoryModel(entity), nil
}

func (r *transactionCategoryRepository) BulkCreate(ctx context.Context, tcs []model.TransactionCategory) ([]model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	if len(tcs) == 0 {
		return []model.TransactionCategory{}, nil
	}

	create := client.TransactionCategory.MapCreateBulk(tcs, func(c *ent.TransactionCategoryCreate, i int) {
		c.SetID(tcs[i].ID).
			SetUserID(tcs[i].UserID).
			SetNillableParentID(tcs[i].ParentID).
			SetName(tcs[i].Name).
			SetNillableSlug(tcs[i].Slug).
			SetIcon(tcs[i].Icon).
			SetColor(tcs[i].Color).
			SetIsDefault(tcs[i].IsDefault).
			SetType(transactioncategory.Type(tcs[i].Type))
	})

	entities, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.TransactionCategory, len(entities))
	for i, e := range entities {
		result[i] = *toTransactionCategoryModel(e)
	}

	return result, nil
}

func (r *transactionCategoryRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.TransactionCategory.Delete().
		Where(transactioncategory.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *transactionCategoryRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.TransactionCategory.Query().
		Where(transactioncategory.IDEQ(ID)).
		Exist(ctx)
}

func (r *transactionCategoryRepository) ExistsByUserIDAndTypeAndName(
	ctx context.Context,
	userID uuid.UUID,
	type_ enum.TransactionCategoryType,
	name string,
) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.TransactionCategory.Query().
		Where(
			transactioncategory.UserIDEQ(userID),
			transactioncategory.TypeEQ(transactioncategory.Type(type_)),
			transactioncategory.Name(name),
		).
		Exist(ctx)
}

func (r *transactionCategoryRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.TransactionCategory.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionCategoryModel(entity), nil
}

func (r *transactionCategoryRepository) GetByIDAndUserID(ctx context.Context, ID uuid.UUID, userID uuid.UUID) (*model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.TransactionCategory.Query().
		Where(
			transactioncategory.IDEQ(ID),
			transactioncategory.UserIDEQ(userID),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionCategoryModel(entity), nil
}

func (r *transactionCategoryRepository) GetByUserIDAndTypeAndName(
	ctx context.Context,
	userID uuid.UUID,
	type_ enum.TransactionCategoryType,
	name string,
) (*model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.TransactionCategory.Query().
		Where(
			transactioncategory.UserIDEQ(userID),
			transactioncategory.TypeEQ(transactioncategory.Type(type_)),
			transactioncategory.Name(name),
		).
		Only(ctx)

	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toTransactionCategoryModel(entity), nil
}

func (r *transactionCategoryRepository) GetByIDsAndUserID(ctx context.Context, IDs []uuid.UUID, userID uuid.UUID) ([]model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.TransactionCategory.Query().
		Where(
			transactioncategory.IDIn(IDs...),
			transactioncategory.UserIDEQ(userID),
		).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.TransactionCategory, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionCategoryModel(e)
	}

	return result, nil
}

func (r *transactionCategoryRepository) GetAllByParentID(ctx context.Context, parentID uuid.UUID) ([]model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.TransactionCategory.Query().
		Where(transactioncategory.ParentIDEQ(parentID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.TransactionCategory, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionCategoryModel(e)
	}

	return result, nil
}

func (r *transactionCategoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	entities, err := client.TransactionCategory.Query().
		Where(transactioncategory.UserIDEQ(userID)).
		All(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.TransactionCategory, len(entities))

	for i, e := range entities {
		result[i] = *toTransactionCategoryModel(e)
	}

	return result, nil
}

func (r *transactionCategoryRepository) Update(ctx context.Context, tc *model.TransactionCategory) (*model.TransactionCategory, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.TransactionCategory.UpdateOneID(tc.ID).
		SetUserID(tc.UserID).
		SetName(tc.Name).
		SetIcon(tc.Icon).
		SetColor(tc.Color).
		SetIsDefault(tc.IsDefault).
		SetType(transactioncategory.Type(tc.Type))

	if tc.ParentID != nil {
		update.SetParentID(*tc.ParentID)
	} else {
		update.ClearParentID()
	}

	if tc.Slug != nil {
		update.SetSlug(*tc.Slug)
	} else {
		update.ClearSlug()
	}

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toTransactionCategoryModel(entity), nil
}
