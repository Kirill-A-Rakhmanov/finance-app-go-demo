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
	"finance-app-go-demo/internal/domain/enum"
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

func newGetCounterpartyReportTestRouter(service *report.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpreport.NewHandler(service)
	r.GET("/report/counterparties", h.GetCounterpartyReport)
	return r
}

func TestHandler_GetCounterpartyReport(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	counterpartyID := uuid.New()
	shopID := uuid.New()

	type mocks struct {
		accounts       *repomock.MockAccountRepository
		counterparties *repomock.MockCounterpartyRepository
		queryService   *querymock.MockReportQueryService
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
			query:   "?type=expense&currency=USD",
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{
						Items: []query.CounterpartyReportItem{
							{Counterparty: &model.Counterparty{ID: shopID, UserID: session.UserID, Name: "Shop"}, Amount: decimal.RequireFromString("600")},
							{Amount: decimal.RequireFromString("400")},
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, float64(1000), body["total"])

				items := body["items"].([]any)
				require.Len(t, items, 2)

				item := items[0].(map[string]any)
				assert.Equal(t, shopID.String(), item["counterpartyId"])
				assert.Equal(t, float64(600), item["amount"])
				assert.Equal(t, float64(60), item["share"])

				withoutCounterparty := items[1].(map[string]any)
				assert.Contains(t, withoutCounterparty, "counterpartyId", "the key is present with a null value")
				assert.Nil(t, withoutCounterparty["counterpartyId"])
				assert.Equal(t, float64(400), withoutCounterparty["amount"])
				assert.Equal(t, float64(40), withoutCounterparty["share"])

				counterparties := body["counterparties"].([]any)
				require.Len(t, counterparties, 1, "only real counterparties are listed")
				assert.Equal(t, shopID.String(), counterparties[0].(map[string]any)["id"])
				assert.Equal(t, "Shop", counterparties[0].(map[string]any)["name"])
			},
		},
		{
			name:    "empty report returns an empty array",
			session: session,
			query:   "?type=income&currency=USD",
			setupMocks: func(m mocks) {
				m.queryService.EXPECT().GetCounterpartyReport(mock.Anything, mock.Anything).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, []any{}, body["items"])
				assert.Equal(t, []any{}, body["counterparties"])
				assert.Equal(t, float64(0), body["total"])
			},
		},
		{
			name:    "query params are converted into criteria",
			session: session,
			query: "?type=income&currency=USD&periodStart=1700000000000&periodEnd=1700086400000" +
				"&accountIds=" + accountID.String() + "&counterpartyIds=" + counterpartyID.String(),
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyUSD}}}, nil)
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{counterpartyID}, session.UserID).
					Return([]model.Counterparty{{ID: counterpartyID}}, nil)
				m.queryService.EXPECT().
					GetCounterpartyReport(mock.Anything, mock.MatchedBy(func(c query.GetCounterpartyReportCriteria) bool {
						return c.UserID == session.UserID &&
							c.TransactionType == enum.TransactionIncome &&
							c.Currency == valueObject.CurrencyUSD &&
							c.PeriodStart != nil && c.PeriodStart.UnixMilli() == 1700000000000 &&
							c.PeriodEnd != nil && c.PeriodEnd.UnixMilli() == 1700086400000 &&
							assert.ObjectsAreEqual([]uuid.UUID{accountID}, c.AccountIDs) &&
							assert.ObjectsAreEqual([]uuid.UUID{counterpartyID}, c.CounterpartyIDs)
					})).
					Return(&query.GetCounterpartyReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			query:      "?type=expense&currency=USD",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid account id returns 400",
			session:    session,
			query:      "?type=expense&currency=USD&accountIds=not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid counterparty id returns 400",
			session:    session,
			query:      "?type=expense&currency=USD&counterpartyIds=not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid period returns 400",
			session:    session,
			query:      "?type=expense&currency=USD&periodStart=abc",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing type returns 422",
			session:    session,
			query:      "?currency=USD",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "transfer type returns 422",
			session:    session,
			query:      "?type=transfer&currency=USD",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing currency returns 422",
			session:    session,
			query:      "?type=expense",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unknown currency returns 400",
			session:    session,
			query:      "?type=expense&currency=XXX",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "foreign account returns 404",
			session: session,
			query:   "?type=expense&currency=USD&accountIds=" + accountID.String(),
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).Return([]model.Account{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "foreign counterparty returns 404",
			session: session,
			query:   "?type=expense&currency=USD&counterpartyIds=" + counterpartyID.String(),
			setupMocks: func(m mocks) {
				m.counterparties.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{counterpartyID}, session.UserID).
					Return([]model.Counterparty{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "account in another currency returns 400",
			session: session,
			query:   "?type=expense&currency=USD&accountIds=" + accountID.String(),
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
				accounts:       repomock.NewMockAccountRepository(t),
				counterparties: repomock.NewMockCounterpartyRepository(t),
				queryService:   querymock.NewMockReportQueryService(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			router := newGetCounterpartyReportTestRouter(report.NewService(m.accounts, nil, m.counterparties, m.queryService, time.UTC), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/counterparties"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
