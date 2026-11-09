package postgres

import (
	"context"
	"finance-app-go-demo/ent"
	"finance-app-go-demo/ent/user"
	"finance-app-go-demo/internal/domain/model"
	"finance-app-go-demo/internal/domain/repository"
	"finance-app-go-demo/internal/domain/valueObject"
	"finance-app-go-demo/pkg/transaction"
	"fmt"

	"github.com/google/uuid"
)

// mapper
func toUserEntity(u *model.User) *ent.User {
	if u == nil {
		return nil
	}

	return &ent.User{
		ID:         u.ID,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Email:      u.Email.ToString(),
		Password:   u.Password,
	}
}

func toUserModel(u *ent.User) (*model.User, error) {
	if u == nil {
		return nil, nil
	}

	email, err := valueObject.NewEmail(u.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to map user %s: %w", u.ID, err)
	}

	return &model.User{
		ID:         u.ID,
		CreatedAt:  u.CreatedAt,
		UpdatedAt:  u.UpdatedAt,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		MiddleName: u.MiddleName,
		Email:      email,
		Password:   u.Password,
	}, nil
}

type userRepository struct {
	client *ent.Client
}

func NewUserRepository(client *ent.Client) repository.UserRepository {
	return &userRepository{client: client}
}

func (r *userRepository) Create(ctx context.Context, u *model.User) (*model.User, error) {
	client := transaction.GetClient(ctx, r.client)

	create := client.User.Create().
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email.ToString()).
		SetPassword(u.Password).
		SetNillableMiddleName(u.MiddleName)

	entity, err := create.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toUserModel(entity)
}

func (r *userRepository) DeleteByID(ctx context.Context, ID uuid.UUID) error {
	client := transaction.GetClient(ctx, r.client)

	_, err := client.User.Delete().
		Where(user.IDEQ(ID)).
		Exec(ctx)

	return err
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.User.Query().
		Where(user.EmailEQ(email)).
		Exist(ctx)
}

func (r *userRepository) ExistsByID(ctx context.Context, ID uuid.UUID) (bool, error) {
	client := transaction.GetClient(ctx, r.client)

	return client.User.Query().
		Where(user.IDEQ(ID)).
		Exist(ctx)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.User.Query().
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toUserModel(entity)
}

func (r *userRepository) GetByID(ctx context.Context, ID uuid.UUID) (*model.User, error) {
	client := transaction.GetClient(ctx, r.client)

	entity, err := client.User.Get(ctx, ID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, nil
		}

		return nil, err
	}

	return toUserModel(entity)
}

func (r *userRepository) Update(ctx context.Context, u *model.User) (*model.User, error) {
	client := transaction.GetClient(ctx, r.client)

	update := client.User.UpdateOneID(u.ID).
		SetFirstName(u.FirstName).
		SetLastName(u.LastName).
		SetEmail(u.Email.ToString()).
		SetPassword(u.Password).
		SetUpdatedAt(u.UpdatedAt)

	if u.MiddleName != nil {
		update.SetMiddleName(*u.MiddleName)
	} else {
		update.ClearMiddleName()
	}

	entity, err := update.Save(ctx)
	if err != nil {
		return nil, err
	}

	return toUserModel(entity)
}
