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
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCategoryDynamicService(t *testing.T, location *time.Location) (*report.Service, *repomock.MockAccountRepository, *repomock.MockTransactionCategoryRepository, *querymock.MockReportQueryService) {
	t.Helper()

	accounts := repomock.NewMockAccountRepository(t)
	categories := repomock.NewMockTransactionCategoryRepository(t)
	queryService := querymock.NewMockReportQueryService(t)

	return report.NewService(accounts, categories, nil, queryService, location), accounts, categories, queryService
}

func TestService_GetCategoryDynamicReport(t *testing.T) {
	userID := uuid.New()
	accountUSD := uuid.New()
	accountEUR := uuid.New()
	categoryID := uuid.New()
	otherCategoryID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	type mocks struct {
		accounts     *repomock.MockAccountRepository
		categories   *repomock.MockTransactionCategoryRepository
		queryService *querymock.MockReportQueryService
	}

	base := report.GetCategoryDynamicReportInput{UserID: userID, Currency: "USD", TransactionCategoryIDs: []uuid.UUID{categoryID}}
	expectCategories := func(m mocks) {
		m.categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, userID).
			Return([]model.TransactionCategory{{ID: categoryID}}, nil)
	}

	tests := []struct {
		name       string
		input      report.GetCategoryDynamicReportInput
		setupMocks func(m mocks)
		wantErr    error
		wantResult func(t *testing.T, result *report.GetCategoryDynamicReportResult)
	}{
		{
			name:  "flattens months into items linked to categories by id",
			input: base,
			setupMocks: func(m mocks) {
				expectCategories(m)
				m.queryService.EXPECT().GetCategoryDynamicReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryDynamicReportResult{Items: []query.CategoryDynamicReportItem{
						{Month: "2026-01", Categories: []query.CategoryDynamicReportCategory{
							{TransactionCategory: model.TransactionCategory{ID: categoryID}, Amount: dec("250.5")},
							{TransactionCategory: model.TransactionCategory{ID: otherCategoryID}, Amount: dec("10")},
						}},
						{Month: "2026-02", Categories: []query.CategoryDynamicReportCategory{}},
						{Month: "2026-03", Categories: []query.CategoryDynamicReportCategory{
							{TransactionCategory: model.TransactionCategory{ID: categoryID}, Amount: dec("7")},
						}},
					}}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryDynamicReportResult) {
				assert.Equal(t, []string{"2026-01", "2026-02", "2026-03"}, result.Months)

				require.Len(t, result.Items, 3)
				assert.Equal(t, "2026-01", result.Items[0].Month)
				assert.Equal(t, categoryID, result.Items[0].TransactionCategoryID)
				assert.True(t, result.Items[0].Amount.Equal(dec("250.5")))
				assert.Equal(t, otherCategoryID, result.Items[1].TransactionCategoryID)
				assert.Equal(t, "2026-03", result.Items[2].Month)
				assert.Equal(t, categoryID, result.Items[2].TransactionCategoryID)

				require.Len(t, result.TransactionCategories, 2, "a category used in several months is listed once")
				assert.Equal(t, categoryID, result.TransactionCategories[0].ID)
				assert.Equal(t, otherCategoryID, result.TransactionCategories[1].ID)
			},
		},
		{
			name:  "empty result is a non-nil list",
			input: base,
			setupMocks: func(m mocks) {
				expectCategories(m)
				m.queryService.EXPECT().GetCategoryDynamicReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryDynamicReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryDynamicReportResult) {
				assert.NotNil(t, result.Months)
				assert.Empty(t, result.Months)
				assert.NotNil(t, result.Items)
				assert.Empty(t, result.Items)
				assert.NotNil(t, result.TransactionCategories)
				assert.Empty(t, result.TransactionCategories)
			},
		},
		{
			name: "criteria are passed through with the service timezone",
			input: report.GetCategoryDynamicReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
				TransactionCategoryIDs: []uuid.UUID{categoryID},
				AccountIDs:             []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				m.categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, userID).
					Return([]model.TransactionCategory{{ID: categoryID}}, nil)
				m.queryService.EXPECT().
					GetCategoryDynamicReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryDynamicReportCriteria) bool {
						return c.UserID == userID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02") == "2026-01-01" &&
							c.PeriodEnd != nil && c.PeriodEnd.Format("2006-01-02") == "2026-03-31" &&
							len(c.TransactionCategoryIDs) == 1 && c.TransactionCategoryIDs[0] == categoryID &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountUSD
					})).
					Return(&query.GetCategoryDynamicReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryDynamicReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name:  "an open period stays open",
			input: base,
			setupMocks: func(m mocks) {
				expectCategories(m)
				m.queryService.EXPECT().
					GetCategoryDynamicReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryDynamicReportCriteria) bool {
						return c.PeriodStart == nil && c.PeriodEnd == nil
					})).
					Return(&query.GetCategoryDynamicReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCategoryDynamicReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign category is not found",
			input: report.GetCategoryDynamicReportInput{
				UserID: userID, Currency: "USD", TransactionCategoryIDs: []uuid.UUID{categoryID},
			},
			setupMocks: func(m mocks) {
				m.categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, userID).
					Return([]model.TransactionCategory{}, nil)
			},
			wantErr: apperror.ErrTransactionCategoryNotFound,
		},
		{
			name: "unknown or foreign account is not found",
			input: report.GetCategoryDynamicReportInput{
				UserID: userID, Currency: "USD", AccountIDs: []uuid.UUID{accountUSD, accountEUR},
				TransactionCategoryIDs: []uuid.UUID{categoryID},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
			},
			wantErr: apperror.ErrAccountNotFound,
		},
		{
			name: "account in another currency is rejected",
			input: report.GetCategoryDynamicReportInput{
				UserID: userID, Currency: "USD", AccountIDs: []uuid.UUID{accountEUR},
				TransactionCategoryIDs: []uuid.UUID{categoryID},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountEUR}, userID).
					Return([]model.Account{newAccount(accountEUR, "EUR")}, nil)
			},
			wantErr: report.ErrAccountCurrencyMismatch,
		},
		{
			name:    "unknown currency",
			input:   report.GetCategoryDynamicReportInput{UserID: userID, Currency: "XXX", TransactionCategoryIDs: []uuid.UUID{categoryID}},
			wantErr: &apperror.Error{Kind: apperror.KindInvalid},
		},
		{
			name:  "query service error is propagated",
			input: base,
			setupMocks: func(m mocks) {
				expectCategories(m)
				m.queryService.EXPECT().GetCategoryDynamicReport(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, accounts, categories, queryService := newCategoryDynamicService(t, time.UTC)
			if tt.setupMocks != nil {
				tt.setupMocks(mocks{accounts: accounts, categories: categories, queryService: queryService})
			}

			result, err := service.GetCategoryDynamicReport(context.Background(), tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Nil(t, result)

				var want *apperror.Error
				if errors.As(tt.wantErr, &want) {
					var got *apperror.Error
					require.ErrorAs(t, err, &got)
					assert.Equal(t, want.Kind, got.Kind)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}

				return
			}

			require.NoError(t, err)
			tt.wantResult(t, result)
		})
	}
}

func TestService_GetCategoryDynamicReport_PeriodIsMadeOfDatesInTheServiceTimezone(t *testing.T) {
	moscow, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	service, _, categories, queryService := newCategoryDynamicService(t, moscow)

	categoryID := uuid.New()
	categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, mock.Anything).
		Return([]model.TransactionCategory{{ID: categoryID}}, nil)

	// 00:30 on March 2nd in Moscow, still March 1st in UTC
	start := time.Date(2026, 3, 1, 21, 30, 0, 0, time.UTC)

	queryService.EXPECT().
		GetCategoryDynamicReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryDynamicReportCriteria) bool {
			return c.Location.String() == "Europe/Moscow" &&
				c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02 15:04") == "2026-03-02 00:00"
		})).
		Return(&query.GetCategoryDynamicReportResult{}, nil)

	_, err = service.GetCategoryDynamicReport(context.Background(), report.GetCategoryDynamicReportInput{
		UserID: uuid.New(), Currency: "USD", PeriodStart: &start, TransactionCategoryIDs: []uuid.UUID{categoryID},
	})

	require.NoError(t, err)
}

func TestService_GetCategoryDynamicReport_Validation(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	tooFar := start.AddDate(report.MaxReportPeriodYears, 0, 1)
	valid := report.GetCategoryDynamicReportInput{UserID: uuid.New(), Currency: "USD", TransactionCategoryIDs: []uuid.UUID{uuid.New()}}

	tests := []struct {
		name      string
		mutate    func(i *report.GetCategoryDynamicReportInput)
		wantField string
	}{
		{"missing user", func(i *report.GetCategoryDynamicReportInput) { i.UserID = uuid.Nil }, "userId"},
		{"missing currency", func(i *report.GetCategoryDynamicReportInput) { i.Currency = "" }, "currency"},
		{"period start after end", func(i *report.GetCategoryDynamicReportInput) { i.PeriodStart, i.PeriodEnd = &start, &end }, "periodStart"},
		{"period over the limit", func(i *report.GetCategoryDynamicReportInput) { i.PeriodStart, i.PeriodEnd = &start, &tooFar }, "periodEnd"},
		{"too many accounts", func(i *report.GetCategoryDynamicReportInput) {
			i.AccountIDs = make([]uuid.UUID, report.MaxReportAccountIDs+1)
		}, "accountIds"},
		{"no categories", func(i *report.GetCategoryDynamicReportInput) { i.TransactionCategoryIDs = nil }, "transactionCategoryIds"},
		{"too many categories", func(i *report.GetCategoryDynamicReportInput) {
			i.TransactionCategoryIDs = make([]uuid.UUID, report.MaxReportCategoryIDs+1)
		}, "transactionCategoryIds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)

			// no expectations: validation must fail before any dependency is called
			service, _, _, _ := newCategoryDynamicService(t, time.UTC)

			result, err := service.GetCategoryDynamicReport(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, result)

			var validationErr *apperror.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Contains(t, validationErr.Fields, tt.wantField)
		})
	}
}
