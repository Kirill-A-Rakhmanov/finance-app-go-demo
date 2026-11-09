package counterparty_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"finance-app-go-demo/internal/application/counterparty"
	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateCounterparty(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      counterparty.CreateCounterpartyInput
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *counterparty.CreateCounterpartyResult)
	}{
		{
			name:  "successful creation",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: "Cafe"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "Cafe").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
						assert.Equal(t, userID, c.UserID)
						c.ID = uuid.New()
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *counterparty.CreateCounterpartyResult) {
				assert.Equal(t, "Cafe", result.Counterparty.Name)
				assert.Equal(t, userID, result.Counterparty.UserID)
			},
		},
		{
			name:  "name is trimmed before the check and the insert",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: "  Cafe "},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "Cafe").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
						c.ID = uuid.New()
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *counterparty.CreateCounterpartyResult) {
				assert.Equal(t, "Cafe", result.Counterparty.Name)
			},
		},
		{
			name:       "validation error - empty fields",
			input:      counterparty.CreateCounterpartyInput{UserID: uuid.Nil, Name: ""},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name:       "validation error - blank name",
			input:      counterparty.CreateCounterpartyInput{UserID: userID, Name: "   "},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name:       "validation error - name too long",
			input:      counterparty.CreateCounterpartyInput{UserID: userID, Name: strings.Repeat("a", model.CounterpartyNameMaxLen+1)},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name:  "name of the maximum length is accepted",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: strings.Repeat("a", model.CounterpartyNameMaxLen)},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, strings.Repeat("a", model.CounterpartyNameMaxLen)).Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
						return c, nil
					})
			},
			wantResult: func(t *testing.T, result *counterparty.CreateCounterpartyResult) {
				assert.Len(t, result.Counterparty.Name, model.CounterpartyNameMaxLen)
			},
		},
		{
			name:  "counterparty with the same name already exists",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: "Cafe"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "Cafe").Return(true, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, counterparty.ErrCounterpartyAlreadyExists)
			},
		},
		{
			name:  "uniqueness check error",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: "Cafe"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "Cafe").Return(false, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "repository create error",
			input: counterparty.CreateCounterpartyInput{UserID: userID, Name: "Cafe"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "Cafe").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(nil, errors.New("insert failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "insert failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			service := counterparty.New(repo)

			result, err := service.CreateCounterparty(context.Background(), tt.input)

			if tt.wantErr != nil {
				tt.wantErr(t, err)
				assert.Nil(t, result)
				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}
