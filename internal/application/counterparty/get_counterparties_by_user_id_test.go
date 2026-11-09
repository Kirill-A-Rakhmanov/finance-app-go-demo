package counterparty_test

import (
	"context"
	"errors"
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

func TestService_GetCounterpartiesByUserID(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		input      counterparty.GetCounterpartiesByUserIDInput
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *counterparty.GetCounterpartiesByUserIDResult)
	}{
		{
			name:  "successful fetch",
			input: counterparty.GetCounterpartiesByUserIDInput{UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.Counterparty{
					{ID: uuid.New(), UserID: userID, Name: "Cafe"},
					{ID: uuid.New(), UserID: userID, Name: "Supermarket"},
				}, nil)
			},
			wantResult: func(t *testing.T, result *counterparty.GetCounterpartiesByUserIDResult) {
				assert.Len(t, result.Counterparties, 2)
			},
		},
		{
			name:  "no counterparties",
			input: counterparty.GetCounterpartiesByUserIDInput{UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, userID).Return([]model.Counterparty{}, nil)
			},
			wantResult: func(t *testing.T, result *counterparty.GetCounterpartiesByUserIDResult) {
				assert.Empty(t, result.Counterparties)
			},
		},
		{
			name:       "validation error - empty user id",
			input:      counterparty.GetCounterpartiesByUserIDInput{UserID: uuid.Nil},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "repository error propagates",
			input: counterparty.GetCounterpartiesByUserIDInput{UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, userID).Return(nil, errors.New("connection refused"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "connection refused")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			service := counterparty.New(repo)

			result, err := service.GetCounterpartiesByUserID(context.Background(), tt.input)

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
