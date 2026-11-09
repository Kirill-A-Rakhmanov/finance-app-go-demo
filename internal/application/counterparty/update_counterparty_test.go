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

func TestService_UpdateCounterparty(t *testing.T) {
	counterpartyID := uuid.New()
	userID := uuid.New()

	stored := func() *model.Counterparty {
		return &model.Counterparty{ID: counterpartyID, UserID: userID, Name: "Old"}
	}
	echoUpdate := func(repo *repomock.MockCounterpartyRepository) {
		repo.EXPECT().Update(mock.Anything, mock.Anything).
			RunAndReturn(func(ctx context.Context, c *model.Counterparty) (*model.Counterparty, error) {
				return c, nil
			})
	}

	tests := []struct {
		name       string
		input      counterparty.UpdateCounterpartyInput
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *counterparty.UpdateCounterpartyResult)
	}{
		{
			name:  "successful rename",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "New").Return(false, nil)
				echoUpdate(repo)
			},
			wantResult: func(t *testing.T, result *counterparty.UpdateCounterpartyResult) {
				assert.Equal(t, "New", result.Counterparty.Name)
			},
		},
		{
			name:  "name is trimmed",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "  New  "},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "New").Return(false, nil)
				echoUpdate(repo)
			},
			wantResult: func(t *testing.T, result *counterparty.UpdateCounterpartyResult) {
				assert.Equal(t, "New", result.Counterparty.Name)
			},
		},
		{
			name:  "unchanged name skips the uniqueness check",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "Old"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				echoUpdate(repo)
			},
			wantResult: func(t *testing.T, result *counterparty.UpdateCounterpartyResult) {
				assert.Equal(t, "Old", result.Counterparty.Name)
			},
		},
		{
			name:       "validation error - empty fields",
			input:      counterparty.UpdateCounterpartyInput{CounterpartyID: uuid.Nil, UserID: uuid.Nil, Name: ""},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "counterpartyId")
				assert.Contains(t, validationErr.Fields, "userId")
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name:       "validation error - name too long",
			input:      counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: strings.Repeat("a", model.CounterpartyNameMaxLen+1)},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "name")
			},
		},
		{
			name:  "counterparty not found or belongs to another user",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrCounterpartyNotFound)
			},
		},
		{
			name:  "renaming to a name that already exists",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "New").Return(true, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, counterparty.ErrCounterpartyAlreadyExists)
			},
		},
		{
			name:  "lookup error",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "uniqueness check error",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "New").Return(false, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "repository update error",
			input: counterparty.UpdateCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID, Name: "New"},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, userID, "New").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).Return(nil, errors.New("update failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "update failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			service := counterparty.New(repo)

			result, err := service.UpdateCounterparty(context.Background(), tt.input)

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
