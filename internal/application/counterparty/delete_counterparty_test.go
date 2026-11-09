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

func TestService_DeleteCounterparty(t *testing.T) {
	counterpartyID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		input      counterparty.DeleteCounterpartyInput
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *counterparty.DeleteCounterpartyResult)
	}{
		{
			name:  "successful deletion",
			input: counterparty.DeleteCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).
					Return(&model.Counterparty{ID: counterpartyID, UserID: userID}, nil)
				repo.EXPECT().DeleteByID(mock.Anything, counterpartyID).Return(nil)
			},
			wantResult: func(t *testing.T, result *counterparty.DeleteCounterpartyResult) {
				assert.Equal(t, "Counterparty deleted successfully", result.Message)
			},
		},
		{
			name:       "validation error - empty fields",
			input:      counterparty.DeleteCounterpartyInput{CounterpartyID: uuid.Nil, UserID: uuid.Nil},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "counterpartyId")
				assert.Contains(t, validationErr.Fields, "userId")
			},
		},
		{
			name:  "counterparty not found or belongs to another user",
			input: counterparty.DeleteCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(nil, nil)
			},
			wantErr: func(t *testing.T, err error) {
				assert.ErrorIs(t, err, apperror.ErrCounterpartyNotFound)
			},
		},
		{
			name:  "lookup error",
			input: counterparty.DeleteCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).Return(nil, errors.New("db is down"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "db is down")
			},
		},
		{
			name:  "repository delete error",
			input: counterparty.DeleteCounterpartyInput{CounterpartyID: counterpartyID, UserID: userID},
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, userID).
					Return(&model.Counterparty{ID: counterpartyID, UserID: userID}, nil)
				repo.EXPECT().DeleteByID(mock.Anything, counterpartyID).Return(errors.New("delete failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "delete failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			service := counterparty.New(repo)

			result, err := service.DeleteCounterparty(context.Background(), tt.input)

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
