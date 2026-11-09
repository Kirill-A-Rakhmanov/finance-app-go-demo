package report_test

import (
	"context"
	"errors"
	"testing"
	"time"

	apperror "finance-app-go-demo/internal/application/error"
	"finance-app-go-demo/internal/application/query"
	querymock "finance-app-go-demo/internal/application/query/mock"
	"finance-app-go-demo/internal/application/report"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func dec(s string) decimal.Decimal {
	return decimal.RequireFromString(s)
}

func newAccount(id uuid.UUID, code valueObject.CurrencyCode) model.Account {
	return model.Account{ID: id, Currency: valueObject.Currency{Code: code}}
}

func TestService_GetCategoryReport(t *testing.T) {
	userID := uuid.New()
	accountUSD := uuid.New()
	accountEUR := uuid.New()
	parentID := uuid.New()
	childID := uuid.New()
	periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	food, rent := uuid.New(), uuid.New()

	base := report.GetCategoryReportInput{
		UserID:          userID,
		TransactionType: enum.TransactionExpense,
		Currency:        "USD",
	}

	tests := []struct {
		name       string
		input      report.GetCategoryReportInput
		setupMocks func(accounts *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService)
		wantErr    error
		wantResult func(t *testing.T, result *report.GetCategoryReportResult)
	}{
		{
			name:  "computes total and shares, keeping the order from the query",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryReportResult{
						Items: []query.CategoryReportItem{
							{TransactionCategory: &model.TransactionCategory{ID: rent}, Amount: dec("600")},
							{TransactionCategory: &model.TransactionCategory{ID: food}, Amount: dec("300")},
							{Amount: dec("100")},
						},
					}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.True(t, result.Total.Equal(dec("1000")), "got %s", result.Total)
				require.Len(t, result.Items, 3)
				require.NotNil(t, result.Items[0].TransactionCategory)
				assert.Equal(t, rent, result.Items[0].TransactionCategory.ID)
				assert.True(t, result.Items[0].Share.Equal(dec("60")), "got %s", result.Items[0].Share)
				require.NotNil(t, result.Items[1].TransactionCategory)
				assert.Equal(t, food, result.Items[1].TransactionCategory.ID)
				assert.True(t, result.Items[1].Share.Equal(dec("30")), "got %s", result.Items[1].Share)
				assert.Nil(t, result.Items[2].TransactionCategory, "the item without a category stays in place")
				assert.True(t, result.Items[2].Amount.Equal(dec("100")))
				assert.True(t, result.Items[2].Share.Equal(dec("10")), "got %s", result.Items[2].Share)
			},
		},
		{
			name:  "shares are rounded to two decimals",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryReportResult{
						Items: []query.CategoryReportItem{
							{TransactionCategory: &model.TransactionCategory{ID: food}, Amount: dec("1")},
							{TransactionCategory: &model.TransactionCategory{ID: rent}, Amount: dec("2")},
						},
					}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.True(t, result.Items[0].Share.Equal(dec("33.33")), "got %s", result.Items[0].Share)
				assert.True(t, result.Items[1].Share.Equal(dec("66.67")), "got %s", result.Items[1].Share)
			},
		},
		{
			name:  "empty report has zero total and a non-nil item list",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.True(t, result.Total.IsZero())
				assert.NotNil(t, result.Items)
				assert.Empty(t, result.Items)
			},
		},
		{
			name:  "only transactions without a category",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryReportResult{Items: []query.CategoryReportItem{{Amount: dec("50")}}}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.True(t, result.Total.Equal(dec("50")))
				require.Len(t, result.Items, 1)
				assert.Nil(t, result.Items[0].TransactionCategory)
				assert.True(t, result.Items[0].Share.Equal(dec("100")))
			},
		},
		{
			name: "criteria are passed through unchanged",
			input: report.GetCategoryReportInput{
				UserID:                 userID,
				TransactionType:        enum.TransactionIncome,
				Currency:               "USD",
				PeriodStart:            &periodStart,
				PeriodEnd:              &periodEnd,
				TransactionCategoryIDs: []uuid.UUID{parentID},
				GroupSubcategories:     true,
				AccountIDs:             []uuid.UUID{accountUSD},
			},
			setupMocks: func(accounts *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{parentID}, userID).Return([]model.TransactionCategory{{ID: parentID}}, nil)
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryReportCriteria) bool {
						return c.UserID == userID &&
							c.TransactionType == enum.TransactionIncome &&
							c.Currency == valueObject.CurrencyUSD &&
							c.PeriodStart != nil && c.PeriodStart.Equal(periodStart) &&
							c.PeriodEnd != nil && c.PeriodEnd.Equal(periodEnd) &&
							len(c.TransactionCategoryIDs) == 1 && c.TransactionCategoryIDs[0] == parentID &&
							c.GroupSubcategories &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountUSD
					})).
					Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.True(t, result.Total.IsZero())
			},
		},
		{
			name:  "without ids the repositories are not touched",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "subcategory is allowed in the filter when not grouping",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				TransactionCategoryIDs: []uuid.UUID{childID},
			},
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{childID}, userID).Return([]model.TransactionCategory{
					{ID: childID, ParentID: &parentID},
				}, nil)
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "subcategory in the filter is rejected when grouping",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				TransactionCategoryIDs: []uuid.UUID{childID}, GroupSubcategories: true,
			},
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{childID}, userID).Return([]model.TransactionCategory{
					{ID: childID, ParentID: &parentID},
				}, nil)
			},
			wantErr: report.ErrSubcategoryFilterNotAllowed,
		},
		{
			name: "duplicate category ids are not mistaken for a missing category",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				TransactionCategoryIDs: []uuid.UUID{parentID, parentID},
			},
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{parentID, parentID}, userID).Return([]model.TransactionCategory{{ID: parentID}}, nil)
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign category is not found",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				TransactionCategoryIDs: []uuid.UUID{parentID, childID},
			},
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{parentID, childID}, userID).Return([]model.TransactionCategory{{ID: parentID}}, nil)
			},
			wantErr: apperror.ErrTransactionCategoryNotFound,
		},
		{
			name: "duplicate account ids are not mistaken for a missing account",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountUSD},
			},
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountUSD}, userID).Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign account is not found",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountEUR},
			},
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
			},
			wantErr: apperror.ErrAccountNotFound,
		},
		{
			name: "account in another currency is rejected",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountEUR},
			},
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).Return([]model.Account{
					newAccount(accountUSD, "USD"), newAccount(accountEUR, "EUR"),
				}, nil)
			},
			wantErr: report.ErrAccountCurrencyMismatch,
		},
		{
			name:    "unknown currency",
			input:   report.GetCategoryReportInput{UserID: userID, TransactionType: enum.TransactionExpense, Currency: "XXX"},
			wantErr: &apperror.Error{Kind: apperror.KindInvalid},
		},
		{
			name: "account repository error is propagated",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD},
			},
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name: "category repository error is propagated",
			input: report.GetCategoryReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				TransactionCategoryIDs: []uuid.UUID{parentID},
			},
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{parentID}, userID).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name:  "query service error is propagated",
			input: base,
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accounts := repomock.NewMockAccountRepository(t)
			categories := repomock.NewMockTransactionCategoryRepository(t)
			queryService := querymock.NewMockReportQueryService(t)
			if tt.setupMocks != nil {
				tt.setupMocks(accounts, categories, queryService)
			}

			service := report.NewService(accounts, categories, nil, queryService, time.UTC)

			result, err := service.GetCategoryReport(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, result)

				var want *apperror.Error
				if errors.As(tt.wantErr, &want) {
					var got *apperror.Error
					require.ErrorAs(t, err, &got)
					assert.Equal(t, want.Kind, got.Kind)
					if want.Message != "" {
						assert.Equal(t, want.Message, got.Message)
					}
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			tt.wantResult(t, result)
		})
	}
}

func TestService_GetCategoryReport_Validation(t *testing.T) {
	userID := uuid.New()
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tooManyIDs := make([]uuid.UUID, report.MaxReportAccountIDs+1)
	for i := range tooManyIDs {
		tooManyIDs[i] = uuid.New()
	}

	valid := report.GetCategoryReportInput{UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD"}

	tests := []struct {
		name   string
		mutate func(i *report.GetCategoryReportInput)
		field  string
	}{
		{"missing user", func(i *report.GetCategoryReportInput) { i.UserID = uuid.Nil }, "userId"},
		{"invalid transaction type", func(i *report.GetCategoryReportInput) { i.TransactionType = "transfer" }, "transactionType"},
		{"empty transaction type", func(i *report.GetCategoryReportInput) { i.TransactionType = "" }, "transactionType"},
		{"empty currency", func(i *report.GetCategoryReportInput) { i.Currency = "" }, "currency"},
		{"period start after end", func(i *report.GetCategoryReportInput) { i.PeriodStart, i.PeriodEnd = &start, &end }, "periodStart"},
		{"too many accounts", func(i *report.GetCategoryReportInput) { i.AccountIDs = tooManyIDs }, "accountIds"},
		{"too many categories", func(i *report.GetCategoryReportInput) { i.TransactionCategoryIDs = tooManyIDs }, "transactionCategoryIds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)

			// no expectations: validation must fail before any dependency is called
			service := report.NewService(
				repomock.NewMockAccountRepository(t),
				repomock.NewMockTransactionCategoryRepository(t),
				nil,
				querymock.NewMockReportQueryService(t),
				time.UTC,
			)

			result, err := service.GetCategoryReport(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, result)

			var validationErr *apperror.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Contains(t, validationErr.Fields, tt.field)
		})
	}
}

func TestService_GetCategoryReport_AcceptsEqualPeriodBounds(t *testing.T) {
	userID := uuid.New()
	day := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	queryService := querymock.NewMockReportQueryService(t)
	queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)

	service := report.NewService(repomock.NewMockAccountRepository(t), repomock.NewMockTransactionCategoryRepository(t), nil, queryService, time.UTC)

	_, err := service.GetCategoryReport(context.Background(), report.GetCategoryReportInput{
		UserID: userID, TransactionType: enum.TransactionIncome, Currency: "USD", PeriodStart: &day, PeriodEnd: &day,
	})
	require.NoError(t, err)
}
