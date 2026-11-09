package transaction_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	querymock "finance-app-go-demo/internal/application/query/mock"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/model"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetTransactions(t *testing.T) {
	userID := uuid.New()
	periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	categoryID := uuid.New()
	otherCategoryID := uuid.New()

	tests := []struct {
		name       string
		input      transaction.GetTransactionsInput
		setupMocks func(queryService *querymock.MockTransactionQueryService)
		wantErr    func(t *testing.T, err error)
		wantResult func(t *testing.T, result *transaction.GetTransactionsResult)
	}{
		{
			name:  "default limit applied when not specified",
			input: transaction.GetTransactionsInput{UserID: userID},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.Limit == transaction.DefaultTransactionsLimit
					})).
					Return(&query.GetTransactionsResult{Total: 2, Transactions: []model.Transaction{{}, {}}}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, transaction.DefaultTransactionsLimit, result.Limit)
				assert.Equal(t, 2, result.Total)
			},
		},
		{
			name:  "explicit limit passed through unchanged",
			input: transaction.GetTransactionsInput{UserID: userID, Limit: 20},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.Limit == 20
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, 20, result.Limit)
			},
		},
		{
			name:  "period boundaries passed through to criteria",
			input: transaction.GetTransactionsInput{UserID: userID, PeriodStart: &periodStart, PeriodEnd: &periodEnd},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.PeriodStart != nil && c.PeriodStart.Equal(periodStart) &&
							c.PeriodEnd != nil && c.PeriodEnd.Equal(periodEnd)
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, 0, result.Total)
			},
		},
		{
			name:  "nil period stays nil in criteria",
			input: transaction.GetTransactionsInput{UserID: userID},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.PeriodStart == nil && c.PeriodEnd == nil
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, 0, result.Total)
			},
		},
		{
			name:  "category ids passed through to criteria",
			input: transaction.GetTransactionsInput{UserID: userID, TransactionCategoryIDs: []uuid.UUID{categoryID, otherCategoryID}},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return assert.ObjectsAreEqual([]uuid.UUID{categoryID, otherCategoryID}, c.TransactionCategoryIDs)
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, 0, result.Total)
			},
		},
		{
			name:  "empty category ids stay empty in criteria",
			input: transaction.GetTransactionsInput{UserID: userID},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return len(c.TransactionCategoryIDs) == 0
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantResult: func(t *testing.T, result *transaction.GetTransactionsResult) {
				assert.Equal(t, 0, result.Total)
			},
		},
		{
			name: "validation error - too many category ids",
			input: transaction.GetTransactionsInput{
				UserID:                 userID,
				TransactionCategoryIDs: make([]uuid.UUID, transaction.MaxTransactionCategoryIDs+1),
			},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "transactionCategoryIds")
			},
		},
		{
			name:  "validation error - limit exceeds max",
			input: transaction.GetTransactionsInput{UserID: userID, Limit: transaction.MaxTransactionsLimit + 1},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "limit")
			},
		},
		{
			name:  "validation error - negative offset",
			input: transaction.GetTransactionsInput{UserID: userID, Offset: -1},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
			},
			wantErr: func(t *testing.T, err error) {
				var validationErr *apperror.ValidationError
				require.True(t, errors.As(err, &validationErr))
				assert.Contains(t, validationErr.Fields, "offset")
			},
		},
		{
			name:  "query service error propagates",
			input: transaction.GetTransactionsInput{UserID: userID},
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().GetTransactions(mock.Anything, mock.Anything).Return(nil, errors.New("query failed"))
			},
			wantErr: func(t *testing.T, err error) {
				assert.EqualError(t, err, "query failed")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryService := querymock.NewMockTransactionQueryService(t)
			tt.setupMocks(queryService)

			service := transaction.NewService(nil, nil, nil, nil, nil, queryService)

			result, err := service.GetTransactions(context.Background(), tt.input)

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
