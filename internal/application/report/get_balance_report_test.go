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

func newBalanceService(t *testing.T, location *time.Location) (*report.Service, *repomock.MockAccountRepository, *querymock.MockReportQueryService) {
	t.Helper()

	accounts := repomock.NewMockAccountRepository(t)
	queryService := querymock.NewMockReportQueryService(t)

	return report.NewService(accounts, nil, nil, queryService, location), accounts, queryService
}

func TestService_GetBalanceReport_Granularity(t *testing.T) {
	userID := uuid.New()
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		days int
		want query.Granularity
	}{
		{"two days is the minimum and goes by day", 2, query.GranularityDay},
		{"thirty days still goes by day", 30, query.GranularityDay},
		{"thirty one days goes by week", 31, query.GranularityWeek},
		{"a year goes by week", 365, query.GranularityWeek},
		{"more than a year goes by month", 366, query.GranularityMonth},
		{"ten years go by month", 3650, query.GranularityMonth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, _, queryService := newBalanceService(t, time.UTC)
			end := start.AddDate(0, 0, tt.days-1)

			queryService.EXPECT().
				GetBalanceReport(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportCriteria) bool {
					return c.Granularity == tt.want
				})).
				Return(&query.GetBalanceReportResult{}, nil)

			result, err := service.GetBalanceReport(context.Background(), report.GetBalanceReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
			})

			require.NoError(t, err)
			assert.Equal(t, tt.want, result.Granularity)
		})
	}
}

func TestService_GetBalanceReport(t *testing.T) {
	userID := uuid.New()
	accountUSD := uuid.New()
	accountEUR := uuid.New()
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	base := report.GetBalanceReportInput{UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end}

	type mocks struct {
		accounts     *repomock.MockAccountRepository
		queryService *querymock.MockReportQueryService
	}

	tests := []struct {
		name       string
		input      report.GetBalanceReportInput
		setupMocks func(m mocks)
		wantErr    error
		wantResult func(t *testing.T, result *report.GetBalanceReportResult)
	}{
		{
			name:  "maps the items of the query",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).
					Return(&query.GetBalanceReportResult{Items: []query.BalanceReportItem{
						{DateKey: "2026-03-09", Balance: dec("100")},
						{DateKey: "2026-03-10", Balance: dec("90.5")},
					}}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetBalanceReportResult) {
				assert.Equal(t, query.GranularityDay, result.Granularity)
				require.Len(t, result.Items, 2)
				assert.Equal(t, "2026-03-09", result.Items[0].DateKey)
				assert.True(t, result.Items[0].Balance.Equal(dec("100")))
				assert.Equal(t, "2026-03-10", result.Items[1].DateKey)
				assert.True(t, result.Items[1].Balance.Equal(dec("90.5")))
			},
		},
		{
			name:  "empty result is a non-nil list",
			input: base,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).Return(&query.GetBalanceReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetBalanceReportResult) {
				assert.NotNil(t, result.Items)
				assert.Empty(t, result.Items)
			},
		},
		{
			name: "criteria are passed through with the service timezone",
			input: report.GetBalanceReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
				AccountIDs: []uuid.UUID{accountUSD},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountUSD}, userID).
					Return([]model.Account{newAccount(accountUSD, "USD")}, nil)
				m.queryService.EXPECT().
					GetBalanceReport(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportCriteria) bool {
						return c.UserID == userID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							c.PeriodStart.Format("2006-01-02") == "2026-03-01" &&
							c.PeriodEnd.Format("2006-01-02") == "2026-03-10" &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountUSD
					})).
					Return(&query.GetBalanceReportResult{}, nil)
			},
			wantResult: func(t *testing.T, result *report.GetBalanceReportResult) {
				assert.NotNil(t, result)
			},
		},
		{
			name: "unknown or foreign account is not found",
			input: report.GetBalanceReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
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
			input: report.GetBalanceReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
				AccountIDs: []uuid.UUID{accountEUR},
			},
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountEUR}, userID).
					Return([]model.Account{newAccount(accountEUR, "EUR")}, nil)
			},
			wantErr: report.ErrAccountCurrencyMismatch,
		},
		{
			name:    "unknown currency",
			input:   report.GetBalanceReportInput{UserID: userID, Currency: "XXX", PeriodStart: &start, PeriodEnd: &end},
			wantErr: &apperror.Error{Kind: apperror.KindInvalid},
		},
		{
			name: "account repository error is propagated",
			input: report.GetBalanceReportInput{
				UserID: userID, Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
				AccountIDs: []uuid.UUID{accountUSD},
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
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))
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

			result, err := service.GetBalanceReport(context.Background(), tt.input)

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

func TestService_GetBalanceReport_PeriodIsMadeOfDatesInTheServiceTimezone(t *testing.T) {
	moscow, err := time.LoadLocation("Europe/Moscow")
	require.NoError(t, err)

	service, _, queryService := newBalanceService(t, moscow)

	// 00:30 on March 2nd and 23:30 on March 9th in Moscow, which are still March 1st and 9th in UTC
	start := time.Date(2026, 3, 1, 21, 30, 0, 0, time.UTC)
	end := time.Date(2026, 3, 9, 20, 30, 0, 0, time.UTC)

	queryService.EXPECT().
		GetBalanceReport(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportCriteria) bool {
			return c.Location.String() == "Europe/Moscow" &&
				c.PeriodStart.Format("2006-01-02 15:04") == "2026-03-02 00:00" &&
				c.PeriodEnd.Format("2006-01-02 15:04") == "2026-03-09 00:00"
		})).
		Return(&query.GetBalanceReportResult{}, nil)

	_, err = service.GetBalanceReport(context.Background(), report.GetBalanceReportInput{
		UserID: uuid.New(), Currency: "USD", PeriodStart: &start, PeriodEnd: &end,
	})

	require.NoError(t, err)
}

func TestService_GetBalanceReport_Validation(t *testing.T) {
	start := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	sameDayEarlier := time.Date(2026, 3, 1, 10, 0, 0, 0, time.UTC)
	sameDayLater := time.Date(2026, 3, 1, 15, 0, 0, 0, time.UTC)
	tooFar := start.AddDate(report.MaxReportPeriodYears, 0, 1)
	valid := report.GetBalanceReportInput{UserID: uuid.New(), Currency: "USD", PeriodStart: &start, PeriodEnd: &end}

	tests := []struct {
		name      string
		mutate    func(i *report.GetBalanceReportInput)
		wantField string
	}{
		{"missing user", func(i *report.GetBalanceReportInput) { i.UserID = uuid.Nil }, "userId"},
		{"missing currency", func(i *report.GetBalanceReportInput) { i.Currency = "" }, "currency"},
		{"start in the future when the end is today", func(i *report.GetBalanceReportInput) {
			future := time.Now().AddDate(0, 0, 5)
			i.PeriodStart, i.PeriodEnd = &future, nil
		}, "periodStart"},
		{"period start after end", func(i *report.GetBalanceReportInput) { i.PeriodStart, i.PeriodEnd = &end, &start }, "periodStart"},
		{"a single day", func(i *report.GetBalanceReportInput) { i.PeriodStart, i.PeriodEnd = &sameDayEarlier, &sameDayLater }, "periodEnd"},
		{"time of day is ignored when checking the order", func(i *report.GetBalanceReportInput) {
			i.PeriodStart, i.PeriodEnd = &sameDayLater, &sameDayEarlier
		}, "periodEnd"},
		{"period over the limit", func(i *report.GetBalanceReportInput) { i.PeriodEnd = &tooFar }, "periodEnd"},
		{"too many accounts", func(i *report.GetBalanceReportInput) {
			i.AccountIDs = make([]uuid.UUID, report.MaxReportAccountIDs+1)
		}, "accountIds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := valid
			tt.mutate(&input)

			// no expectations: validation must fail before any dependency is called
			service, _, _ := newBalanceService(t, time.UTC)

			result, err := service.GetBalanceReport(context.Background(), input)

			require.Error(t, err)
			assert.Nil(t, result)

			var validationErr *apperror.ValidationError
			require.ErrorAs(t, err, &validationErr)
			assert.Contains(t, validationErr.Fields, tt.wantField)
		})
	}
}

func TestService_GetBalanceReport_DefaultPeriod(t *testing.T) {
	userID := uuid.New()
	accountID := uuid.New()
	today := time.Now().UTC()

	daysAgo := func(days int) time.Time {
		year, month, day := today.Date()
		return time.Date(year, month, day-days, 0, 0, 0, 0, time.UTC)
	}
	dateOf := func(t time.Time) string { return t.Format("2006-01-02") }

	tests := []struct {
		name            string
		input           report.GetBalanceReportInput
		firstActivity   *time.Time
		skipStartLookup bool
		wantStart       string
		wantEnd         string
		wantGranularity query.Granularity
	}{
		{
			name:            "end defaults to today",
			input:           report.GetBalanceReportInput{PeriodStart: ptr(daysAgo(10))},
			skipStartLookup: true,
			wantStart:       dateOf(daysAgo(10)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityDay,
		},
		{
			name:            "start defaults to the first activity",
			firstActivity:   ptr(daysAgo(40)),
			wantStart:       dateOf(daysAgo(40)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityWeek,
		},
		{
			name:            "a long history is shown by month",
			firstActivity:   ptr(daysAgo(1000)),
			wantStart:       dateOf(daysAgo(1000)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityMonth,
		},
		{
			name:            "without any activity the period is the minimum",
			wantStart:       dateOf(daysAgo(1)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityDay,
		},
		{
			name:            "activity starting today still gives the minimum period",
			firstActivity:   ptr(daysAgo(0)),
			wantStart:       dateOf(daysAgo(1)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityDay,
		},
		{
			name:            "an end before the first activity gives the minimum period before it",
			input:           report.GetBalanceReportInput{PeriodEnd: ptr(daysAgo(100))},
			firstActivity:   ptr(daysAgo(50)),
			wantStart:       dateOf(daysAgo(101)),
			wantEnd:         dateOf(daysAgo(100)),
			wantGranularity: query.GranularityDay,
		},
		{
			name:            "activity older than the limit is cut",
			firstActivity:   ptr(today.AddDate(-report.MaxReportPeriodYears-1, 0, 0)),
			wantStart:       dateOf(daysAgo(0).AddDate(-report.MaxReportPeriodYears, 0, 0)),
			wantEnd:         dateOf(daysAgo(0)),
			wantGranularity: query.GranularityMonth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, accounts, queryService := newBalanceService(t, time.UTC)

			input := tt.input
			input.UserID = userID
			input.Currency = "USD"
			input.AccountIDs = []uuid.UUID{accountID}

			accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, userID).
				Return([]model.Account{newAccount(accountID, "USD")}, nil)

			if !tt.skipStartLookup {
				queryService.EXPECT().
					GetBalanceReportStart(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportStartCriteria) bool {
						return c.UserID == userID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							len(c.AccountIDs) == 1 && c.AccountIDs[0] == accountID
					})).
					Return(&query.GetBalanceReportStartResult{Date: tt.firstActivity}, nil)
			}

			queryService.EXPECT().
				GetBalanceReport(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportCriteria) bool {
					return c.PeriodStart.Format("2006-01-02") == tt.wantStart &&
						c.PeriodEnd.Format("2006-01-02") == tt.wantEnd &&
						c.Granularity == tt.wantGranularity
				})).
				Return(&query.GetBalanceReportResult{}, nil)

			result, err := service.GetBalanceReport(context.Background(), input)

			require.NoError(t, err)
			assert.Equal(t, tt.wantGranularity, result.Granularity)
		})
	}
}

func TestService_GetBalanceReport_StartLookupErrorIsPropagated(t *testing.T) {
	service, _, queryService := newBalanceService(t, time.UTC)

	queryService.EXPECT().GetBalanceReportStart(mock.Anything, mock.Anything).Return(nil, errors.New("db down"))

	result, err := service.GetBalanceReport(context.Background(), report.GetBalanceReportInput{
		UserID: uuid.New(), Currency: "USD",
	})

	require.EqualError(t, err, "db down")
	assert.Nil(t, result)
}

func ptr[T any](v T) *T {
	return &v
}
