package report_test

import (
	"encoding/json"
	"fmt"
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

func newGetBalanceReportTestRouter(service *report.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpreport.NewHandler(service)
	r.GET("/report/balance", h.GetBalanceReport)
	return r
}

func TestHandler_GetBalanceReport(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	start := time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC).UnixMilli()
	end := time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC).UnixMilli()
	period := fmt.Sprintf("&periodStart=%d&periodEnd=%d", start, end)

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
			query:   "?currency=USD" + period,
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).
					Return(&query.GetBalanceReportResult{Items: []query.BalanceReportItem{
						{DateKey: "2026-03-01", Balance: decimal.RequireFromString("100")},
						{DateKey: "2026-03-02", Balance: decimal.RequireFromString("90.5")},
					}}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "day", body["granularity"])

				items := body["items"].([]any)
				require.Len(t, items, 2)
				assert.Equal(t, "2026-03-01", items[0].(map[string]any)["dateKey"])
				assert.Equal(t, float64(100), items[0].(map[string]any)["balance"])
				assert.Equal(t, "2026-03-02", items[1].(map[string]any)["dateKey"])
				assert.Equal(t, 90.5, items[1].(map[string]any)["balance"])
			},
		},
		{
			name:    "granularity follows the length of the period",
			session: session,
			query: fmt.Sprintf("?currency=USD&periodStart=%d&periodEnd=%d",
				time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli(),
				time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC).UnixMilli()),
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).Return(&query.GetBalanceReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "week", body["granularity"])
				assert.Equal(t, []any{}, body["items"])
			},
		},
		{
			name:    "query params are converted into criteria",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String() + period,
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyUSD}}}, nil)
				m.queryService.EXPECT().
					GetBalanceReport(mock.Anything, mock.MatchedBy(func(c query.GetBalanceReportCriteria) bool {
						return c.UserID == session.UserID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Granularity == query.GranularityDay &&
							c.PeriodStart.Format("2006-01-02") == "2026-03-01" &&
							c.PeriodEnd.Format("2006-01-02") == "2026-03-10" &&
							assert.ObjectsAreEqual([]uuid.UUID{accountID}, c.AccountIDs)
					})).
					Return(&query.GetBalanceReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			query:      "?currency=USD" + period,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid account id returns 400",
			session:    session,
			query:      "?currency=USD&accountIds=not-a-uuid" + period,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid period returns 400",
			session:    session,
			query:      "?currency=USD&periodStart=abc&periodEnd=1",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "missing period uses the defaults",
			session: session,
			query:   "?currency=USD",
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetBalanceReportStart(mock.Anything, mock.Anything).
					Return(&query.GetBalanceReportStartResult{}, nil)
				m.queryService.EXPECT().GetBalanceReport(mock.Anything, mock.Anything).
					Return(&query.GetBalanceReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "day", body["granularity"], "without activity the period is the minimum of two days")
			},
		},
		{
			name:       "single day period returns 422",
			session:    session,
			query:      fmt.Sprintf("?currency=USD&periodStart=%d&periodEnd=%d", start, start),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "period start after end returns 422",
			session:    session,
			query:      fmt.Sprintf("?currency=USD&periodStart=%d&periodEnd=%d", end, start),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing currency returns 422",
			session:    session,
			query:      "?" + period[1:],
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unknown currency returns 400",
			session:    session,
			query:      "?currency=XXX" + period,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "foreign account returns 404",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String() + period,
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).Return([]model.Account{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "account in another currency returns 400",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String() + period,
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

			router := newGetBalanceReportTestRouter(report.NewService(m.accounts, nil, nil, m.queryService, time.UTC), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/balance"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
