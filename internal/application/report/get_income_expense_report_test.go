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

func TestService_GetIncomeExpenseReport(t *testing.T) {
	userID := uuid.New()
	accountUSD := uuid.New()
	accountEUR := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	base := report.GetIncomeExpenseReportInput{UserID: userID, Currency: "USD"}

	type mocks struct {
		accounts     *repomock.MockAccountRepository
		queryService *querymock.MockReportQueryService
	}

	tests := []struct {
		name       string
		input      report.GetIncomeExpenseReportInput
		setupMocks func(m mocks)
		wantErr    error
		wantResult func(t *testing.T, result *report.GetIncomeExpenseReportResult)
	}{
		{
			name:  "maps the items of the query",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetIncomeExpenseReport(mock.Anything, mock.Anything).
					Return(&query.GetIncomeExpenseReportResult{Items: []query.IncomeExpenseReportItem{
						{Month: "2026-01", Income: dec("1000"), Expense: dec("250.5")},
						{Month: "2026-02", Income: dec("0"), Expense: dec("0")},
					}}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetIncomeExpenseReportResult) {
				require.Len(t, result.Items, 2)
				assert.Equal(t, "2026-01", result.Items[0].Month)
				assert.True(t, result.Items[0].Income.Equal(dec("1000")))
				assert.True(t, result.Items[0].Expense.Equal(dec("250.5")))
				assert.Equal(t, "2026-02", result.Items[1].Month)
				assert.True(t, result.Items[1].Income.IsZero())
			},
		},
		{
			name:  "empty result is a non-nil list",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetIncomeExpenseReport(mock.Anything, mock.Anything).
					Return(&query.GetIncomeExpenseReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetIncomeExpenseReportResult) {
				assert.NotNil(t, result.Items)
				assert.Empty(t, result.Items)
			},
		},
		{
			name: "criteria are passed through with the service timezone",
			input: report.GetIncomeExpenseReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
				AccountIDs: []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				m.queryService.EXPECT().
					GetIncomeExpenseReport(mock.Anything, mock.MatchedBy(func(c query.GetIncomeExpenseReportCriteria) bool {
						return c.UserID == userID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02") == "2026-01-01" &&
							c.PeriodEnd != nil && c.PeriodEnd.Format("2006-01-02") == "2026-03-31" &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountUSD
					})).
					Return(&query.GetIncomeExpenseReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetIncomeExpenseReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name:  "an open period stays open",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().
					GetIncomeExpenseReport(mock.Anything, mock.MatchedBy(func(c query.GetIncomeExpenseReportCriteria) bool {
						return c.PeriodStart == nil && c.PeriodEnd == nil
					})).
					Return(&query.GetIncomeExpenseReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetIncomeExpenseReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign account is not found",
			input: report.GetIncomeExpenseReportInput{
				UserID: userID, Currency: "USD", AccountIDs: []uuid.UUID{accountUSD, accountEUR},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
			},
			wantErr: apperror.ErrAccountNotFound,
		},
		{
			name: "account in another currency is rejected",
			input: report.GetIncomeExpenseReportInput{
				UserID: userID, Currency: "USD", AccountIDs: []uuid.UUID{accountEUR},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountEUR}, userID).
					Return([]model.Account{newAccount(accountEUR, "EUR")}, nil)
			},
			wantErr: report.ErrAccountCurrencyMismatch,
		},
		{
			name:    "unknown currency",
			input:   report.GetIncomeExpenseReportInput{UserID: userID, Currency: "XXX"},
			wantErr: &apperror.Error{Kind: apperror.KindInvalid},
		},
		{
			name: "account repository error is propagated",
			input: report.GetIncomeExpenseReportInput{
				UserID: userID, Currency: "USD", AccountIDs: []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name:  "query service error is propagated",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetIncomeExpenseReport(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, accounts, queryService := newBalanceService(t, time.UTC)
			if tt.setupMocks != nil {
				tt.setupMocks(mocks{accounts: accounts, queryService: queryService})
			}

			result, err := service.GetIncomeExpenseReport(context.Background(), tt.input)

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

func TestService_GetIncomeExpenseReport_PeriodIsMadeOfDatesInTheServiceTimezone(t *testing.T) {
	moscow, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	service, _, queryService := newBalanceService(t, moscow)

	// 00:30 on March 2nd in Moscow, still March 1st in UTC
	start := time.Date(2026, 3, 1, 21, 30, 0, 0, time.UTC)

	queryService.EXPECT().
		GetIncomeExpenseReport(mock.Anything, mock.MatchedBy(func(c query.GetIncomeExpenseReportCriteria) bool {
			return c.Location.String() == "Europe/Moscow" &&
				c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02 15:04") == "2026-03-02 00:00"
		})).
		Return(&query.GetIncomeExpenseReportResult{}, nil)

	_, err = service.GetIncomeExpenseReport(context.Background(), report.GetIncomeExpenseReportInput{
		UserID: uuid.New(), Currency: "USD", PeriodStart: &start,
	})

	require.NoError(t, err)
}

func TestService_GetIncomeExpenseReport_Validation(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	tooFar := start.AddDate(report.MaxReportPeriodYears, 0, 1)
	valid := report.GetIncomeExpenseReportInput{UserID: uuid.New(), Currency: "USD"}

	tests := []struct {
		name      string
		mutate    func(i *report.GetIncomeExpenseReportInput)
		wantField string
	}{
		{"missing user", func(i *report.GetIncomeExpenseReportInput) { i.UserID = uuid.Nil }, "userId"},
		{"missing currency", func(i *report.GetIncomeExpenseReportInput) { i.Currency = "" }, "currency"},
		{"period start after end", func(i *report.GetIncomeExpenseReportInput) { i.PeriodStart, i.PeriodEnd = &start, &end }, "periodStart"},
		{"period over the limit", func(i *report.GetIncomeExpenseReportInput) { i.PeriodStart, i.PeriodEnd = &start, &tooFar }, "periodEnd"},
		{"too many accounts", func(i *report.GetIncomeExpenseReportInput) {
			i.AccountIDs = make([]uuid.UUID, report.MaxReportAccountIDs+1)
		}, "accountIds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)

			// no expectations: validation must fail before any dependency is called
			service, _, _ := newBalanceService(t, time.UTC)

			result, err := service.GetIncomeExpenseReport(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, result)

			var validationErr *apperror.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Contains(t, validationErr.Fields, tt.wantField)
		})
	}
}
