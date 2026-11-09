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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_GetCounterpartyReport(t *testing.T) {
	userID := uuid.New()
	accountUSD := uuid.New()
	accountEUR := uuid.New()
	periodStart := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	cafe, shop := uuid.New(), uuid.New()

	base := report.GetCounterpartyReportInput{
		UserID:          userID,
		TransactionType: enum.TransactionExpense,
		Currency:        "USD",
	}

	type mocks struct {
		accounts       *repomock.MockAccountRepository
		counterparties *repomock.MockCounterpartyRepository
		queryService   *querymock.MockReportQueryService
	}

	tests := []struct {
		name       string
		input      report.GetCounterpartyReportInput
		setupMocks func(m mocks)
		wantErr    error
		wantResult func(t *testing.T, result *report.GetCounterpartyReportResult)
	}{
		{
			name:  "computes total and shares, keeping the order from the query",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{
						Items: []query.CounterpartyReportItem{
							{Counterparty: &model.Counterparty{ID: shop}, Amount: dec("600")},
							{Counterparty: &model.Counterparty{ID: cafe}, Amount: dec("300")},
							{Amount: dec("100")},
						},
					}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.True(t, result.Total.Equal(dec("1000")), "got %s", result.Total)
				require.Len(t, result.Items, 3)
				require.NotNil(t, result.Items[0].Counterparty)
				assert.Equal(t, shop, result.Items[0].Counterparty.ID)
				assert.True(t, result.Items[0].Share.Equal(dec("60")), "got %s", result.Items[0].Share)
				require.NotNil(t, result.Items[1].Counterparty)
				assert.Equal(t, cafe, result.Items[1].Counterparty.ID)
				assert.True(t, result.Items[1].Share.Equal(dec("30")), "got %s", result.Items[1].Share)
				assert.Nil(t, result.Items[2].Counterparty, "the item without a counterparty stays in place")
				assert.True(t, result.Items[2].Amount.Equal(dec("100")))
				assert.True(t, result.Items[2].Share.Equal(dec("10")), "got %s", result.Items[2].Share)
			},
		},
		{
			name:  "shares are rounded to two decimals",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{
						Items: []query.CounterpartyReportItem{
							{Counterparty: &model.Counterparty{ID: cafe}, Amount: dec("1")},
							{Counterparty: &model.Counterparty{ID: shop}, Amount: dec("2")},
						},
					}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.True(t, result.Items[0].Share.Equal(dec("33.33")), "got %s", result.Items[0].Share)
				assert.True(t, result.Items[1].Share.Equal(dec("66.67")), "got %s", result.Items[1].Share)
			},
		},
		{
			name:  "empty report has zero total and a non-nil item list",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.True(t, result.Total.IsZero())
				assert.NotNil(t, result.Items)
				assert.Empty(t, result.Items)
			},
		},
		{
			name:  "only transactions without a counterparty",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{Items: []query.CounterpartyReportItem{{Amount: dec("50")}}}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.True(t, result.Total.Equal(dec("50")))
				require.Len(t, result.Items, 1)
				assert.Nil(t, result.Items[0].Counterparty)
				assert.True(t, result.Items[0].Share.Equal(dec("100")))
			},
		},
		{
			name: "criteria are passed through unchanged",
			input: report.GetCounterpartyReportInput{
				UserID:          userID,
				TransactionType: enum.TransactionIncome,
				Currency:        "USD",
				PeriodStart:     &periodStart,
				PeriodEnd:       &periodEnd,
				CounterpartyIDs: []uuid.UUID{cafe},
				AccountIDs:      []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{cafe}, userID).
					Return([]model.Counterparty{{ID: cafe}}, nil)
				m.queryService.EXPECT().
					GetCounterpartyReport(mock.Anything, mock.MatchedBy(func(c query.GetCounterpartyReportCriteria) bool {
						return c.UserID == userID &&
							c.TransactionType == enum.TransactionIncome &&
							c.Currency == valueObject.CurrencyUSD &&
							c.PeriodStart != nil && c.PeriodStart.Equal(periodStart) &&
							c.PeriodEnd != nil && c.PeriodEnd.Equal(periodEnd) &&
							len(c.CounterpartyIDs) == 1 && c.CounterpartyIDs[0] == cafe &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountUSD
					})).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.True(t, result.Total.IsZero())
			},
		},
		{
			name: "duplicate counterparty ids are not mistaken for a missing counterparty",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				CounterpartyIDs: []uuid.UUID{cafe, cafe},
			},
			setupMocks: func(m mocks) {
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{cafe, cafe}, userID).
					Return([]model.Counterparty{{ID: cafe}}, nil)
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign counterparty is not found",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				CounterpartyIDs: []uuid.UUID{cafe, shop},
			},
			setupMocks: func(m mocks) {
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{cafe, shop}, userID).
					Return([]model.Counterparty{{ID: cafe}}, nil)
			},
			wantErr: apperror.ErrCounterpartyNotFound,
		},
		{
			name: "duplicate account ids are not mistaken for a missing account",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountUSD}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetCounterpartyReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign account is not found",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountEUR},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
			},
			wantErr: apperror.ErrAccountNotFound,
		},
		{
			name: "account in another currency is rejected",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD, accountEUR},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD, accountEUR}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD"), newAccount(accountEUR, "EUR")}, nil)
			},
			wantErr: report.ErrAccountCurrencyMismatch,
		},
		{
			name:    "unknown currency",
			input:   report.GetCounterpartyReportInput{UserID: userID, TransactionType: enum.TransactionExpense, Currency: "XXX"},
			wantErr: &apperror.Error{Kind: apperror.KindInvalid},
		},
		{
			name: "account repository error is propagated",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				AccountIDs: []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name: "counterparty repository error is propagated",
			input: report.GetCounterpartyReportInput{
				UserID: userID, TransactionType: enum.TransactionExpense, Currency: "USD",
				CounterpartyIDs: []uuid.UUID{cafe},
			},
			setupMocks: func(m mocks) {
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{cafe}, userID).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
		{
			name:  "query service error is propagated",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
			},
			wantErr: errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks{
				accounts:       repomock.NewMockAccountRepository(t),
				counterparties: repomock.NewMockCounterpartyRepository(t),
				queryService:   querymock.NewMockReportQueryService(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			service := report.NewService(m.accounts, nil, m.counterparties, m.queryService, time.UTC)

			result, err := service.GetCounterpartyReport(context.Background(), tt.input)

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
			tt.wantResult(t, result)
		})
	}
}

func TestService_GetCounterpartyReport_Validation(t *testing.T) {
	valid := report.GetCounterpartyReportInput{
		UserID:          uuid.New(),
		TransactionType: enum.TransactionExpense,
		Currency:        "USD",
	}
	start := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		mutate    func(i *report.GetCounterpartyReportInput)
		wantField string
	}{
		{"missing user", func(i *report.GetCounterpartyReportInput) { i.UserID = uuid.Nil }, "userId"},
		{"transfer type", func(i *report.GetCounterpartyReportInput) { i.TransactionType = "transfer" }, "transactionType"},
		{"missing currency", func(i *report.GetCounterpartyReportInput) { i.Currency = "" }, "currency"},
		{"period start after end", func(i *report.GetCounterpartyReportInput) { i.PeriodStart, i.PeriodEnd = &start, &end }, "periodStart"},
		{"too many accounts", func(i *report.GetCounterpartyReportInput) {
			i.AccountIDs = make([]uuid.UUID, report.MaxReportAccountIDs+1)
		}, "accountIds"},
		{"too many counterparties", func(i *report.GetCounterpartyReportInput) {
			i.CounterpartyIDs = make([]uuid.UUID, report.MaxReportCounterpartyIDs+1)
		}, "counterpartyIds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)

			// no expectations: validation must fail before any dependency is called
			service := report.NewService(
				repomock.NewMockAccountRepository(t),
				nil,
				repomock.NewMockCounterpartyRepository(t),
				querymock.NewMockReportQueryService(t),
				time.UTC,
			)

			result, err := service.GetCounterpartyReport(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, result)

			var validationErr *apperror.ValidationError
			require.True(t, errors.As(err, &validationErr), "got %v", err)
			assert.Contains(t, validationErr.Fields, tt.wantField)
		})
	}

}
