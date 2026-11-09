package report_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httpreport "finance-app-go-demo/internal/adapter/http/report"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/query"
	querymock "finance-app-go-demo/internal/application/query/mock"
	"finance-app-go-demo/internal/application/report"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetIncomeExpenseReportTestRouter(service *report.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpreport.NewHandler(service)
	r.GET("/report/income-expense", h.GetIncomeExpenseReport)
	return r
}

func TestHandler_GetIncomeExpenseReport(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()

	type mocks struct {
		accounts     *repomock.MockAccountRepository
		queryService *querymock.MockReportQueryService
	}

	tests := []struct {
		name       string
		session    *request.Session
		query      string
		setupMocks func(m mocks)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful report",
			session: session,
			query:   "?currency=USD",
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetIncomeExpenseReport(mock.Anything, mock.Anything).
					Return(&query.GetIncomeExpenseReportResult{Items: []query.IncomeExpenseReportItem{
						{Month: "2026-01", Income: decimal.RequireFromString("1000"), Expense: decimal.RequireFromString("250.5")},
						{Month: "2026-02", Income: decimal.Zero, Expense: decimal.Zero},
					}}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

				items := body["items"].([]any)
				require.Len(t, items, 2)

				first := items[0].(map[string]any)
				assert.Equal(t, "2026-01", first["month"])
				assert.Equal(t, float64(1000), first["income"])
				assert.Equal(t, 250.5, first["expense"])

				empty := items[1].(map[string]any)
				assert.Equal(t, "2026-02", empty["month"])
				assert.Equal(t, float64(0), empty["income"], "zero amounts are present, not omitted")
				assert.Equal(t, float64(0), empty["expense"])
			},
		},
		{
			name:    "empty report returns an empty array",
			session: session,
			query:   "?currency=USD",
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetIncomeExpenseReport(mock.Anything, mock.Anything).
					Return(&query.GetIncomeExpenseReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, []any{}, body["items"])
			},
		},
		{
			name:    "query params are converted into criteria",
			session: session,
			query: "?currency=USD&periodStart=1700000000000&periodEnd=1702592000000" +
				"&accountIds=" + accountID.String(),
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyUSD}}}, nil)
				m.queryService.EXPECT().
					GetIncomeExpenseReport(mock.Anything, mock.MatchedBy(func(c query.GetIncomeExpenseReportCriteria) bool {
						return c.UserID == session.UserID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02") == "2023-11-14" &&
							c.PeriodEnd != nil && c.PeriodEnd.Format("2006-01-02") == "2023-12-14" &&
							assert.ObjectsAreEqual([]uuid.UUID{accountID}, c.AccountIDs)
					})).
					Return(&query.GetIncomeExpenseReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			query:      "?currency=USD",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid account id returns 400",
			session:    session,
			query:      "?currency=USD&accountIds=not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid period returns 400",
			session:    session,
			query:      "?currency=USD&periodStart=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "period start after end returns 422",
			session:    session,
			query:      "?currency=USD&periodStart=1702592000000&periodEnd=1700000000000",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing currency returns 422",
			session:    session,
			query:      "",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unknown currency returns 400",
			session:    session,
			query:      "?currency=XXX",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "foreign account returns 404",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String(),
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).Return([]model.Account{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "account in another currency returns 400",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String(),
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyEUR}}}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks{
				accounts:     repomock.NewMockAccountRepository(t),
				queryService: querymock.NewMockReportQueryService(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			router := newGetIncomeExpenseReportTestRouter(report.NewService(m.accounts, nil, nil, m.queryService, time.UTC), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/income-expense"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
