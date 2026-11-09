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

func newGetCategoryDynamicReportTestRouter(service *report.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpreport.NewHandler(service)
	r.GET("/report/category-dynamic", h.GetCategoryDynamicReport)
	return r
}

func TestHandler_GetCategoryDynamicReport(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	foodID := uuid.New()
	rentID := uuid.New()
	categoriesQuery := "&transactionCategoryIds=" + foodID.String()

	type mocks struct {
		accounts     *repomock.MockAccountRepository
		categories   *repomock.MockTransactionCategoryRepository
		queryService *querymock.MockReportQueryService
	}

	expectCategories := func(m mocks, ids ...uuid.UUID) {
		found := make([]model.TransactionCategory, len(ids))
		for i, id := range ids {
			found[i] = model.TransactionCategory{ID: id}
		}

		m.categories.EXPECT().GetByIDsAndUserID(mock.Anything, ids, session.UserID).Return(found, nil)
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
			name:    "successful report links items to categories by id",
			session: session,
			query:   "?currency=USD" + categoriesQuery + "&transactionCategoryIds=" + rentID.String(),
			setupMocks: func(m mocks) {
				expectCategories(m, foodID, rentID)
				m.queryService.EXPECT().GetCategoryDynamicReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryDynamicReportResult{Items: []query.CategoryDynamicReportItem{
						{Month: "2026-01", Categories: []query.CategoryDynamicReportCategory{
							{TransactionCategory: model.TransactionCategory{ID: rentID, Name: "Rent"}, Amount: decimal.RequireFromString("100")},
							{TransactionCategory: model.TransactionCategory{ID: foodID, Name: "Food"}, Amount: decimal.RequireFromString("25.5")},
						}},
						{Month: "2026-02", Categories: []query.CategoryDynamicReportCategory{}},
						{Month: "2026-03", Categories: []query.CategoryDynamicReportCategory{
							{TransactionCategory: model.TransactionCategory{ID: foodID, Name: "Food"}, Amount: decimal.RequireFromString("20")},
						}},
					}}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

				assert.Equal(t, []any{"2026-01", "2026-02", "2026-03"}, body["months"])

				items := body["items"].([]any)
				require.Len(t, items, 3)

				first := items[0].(map[string]any)
				assert.Equal(t, "2026-01", first["month"])
				assert.Equal(t, rentID.String(), first["transactionCategoryId"])
				assert.Equal(t, float64(100), first["amount"])

				assert.Equal(t, 25.5, items[1].(map[string]any)["amount"])
				assert.Equal(t, "2026-03", items[2].(map[string]any)["month"])

				categories := body["transactionCategories"].([]any)
				require.Len(t, categories, 2, "each category is listed once")
				assert.Equal(t, rentID.String(), categories[0].(map[string]any)["id"])
				assert.Equal(t, foodID.String(), categories[1].(map[string]any)["id"])
			},
		},
		{
			name:    "empty report returns empty arrays",
			session: session,
			query:   "?currency=USD" + categoriesQuery,
			setupMocks: func(m mocks) {
				expectCategories(m, foodID)
				m.queryService.EXPECT().GetCategoryDynamicReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryDynamicReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, []any{}, body["months"])
				assert.Equal(t, []any{}, body["items"])
				assert.Equal(t, []any{}, body["transactionCategories"])
			},
		},
		{
			name:    "query params are converted into criteria",
			session: session,
			query: "?currency=USD&periodStart=1700000000000&periodEnd=1702592000000" +
				"&accountIds=" + accountID.String() + categoriesQuery,
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyUSD}}}, nil)
				expectCategories(m, foodID)
				m.queryService.EXPECT().
					GetCategoryDynamicReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryDynamicReportCriteria) bool {
						return c.UserID == session.UserID &&
							c.Currency == valueObject.CurrencyUSD &&
							c.Location.String() == "UTC" &&
							c.PeriodStart != nil && c.PeriodStart.Format("2006-01-02") == "2023-11-14" &&
							c.PeriodEnd != nil && c.PeriodEnd.Format("2006-01-02") == "2023-12-14" &&
							assert.ObjectsAreEqual([]uuid.UUID{accountID}, c.AccountIDs) &&
							assert.ObjectsAreEqual([]uuid.UUID{foodID}, c.TransactionCategoryIDs)
					})).
					Return(&query.GetCategoryDynamicReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			query:      "?currency=USD" + categoriesQuery,
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid account id returns 400",
			session:    session,
			query:      "?currency=USD&accountIds=not-a-uuid" + categoriesQuery,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid category id returns 400",
			session:    session,
			query:      "?currency=USD&transactionCategoryIds=not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid period returns 400",
			session:    session,
			query:      "?currency=USD&periodStart=abc" + categoriesQuery,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "period start after end returns 422",
			session:    session,
			query:      "?currency=USD&periodStart=1702592000000&periodEnd=1700000000000" + categoriesQuery,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing categories returns 422",
			session:    session,
			query:      "?currency=USD",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "missing currency returns 422",
			session:    session,
			query:      "?" + categoriesQuery[1:],
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "unknown currency returns 400",
			session:    session,
			query:      "?currency=XXX" + categoriesQuery,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "foreign category returns 404",
			session: session,
			query:   "?currency=USD" + categoriesQuery,
			setupMocks: func(m mocks) {
				m.categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{foodID}, session.UserID).
					Return([]model.TransactionCategory{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "foreign account returns 404",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String() + categoriesQuery,
			setupMocks: func(m mocks) {
				m.accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).Return([]model.Account{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "account in another currency returns 400",
			session: session,
			query:   "?currency=USD&accountIds=" + accountID.String() + categoriesQuery,
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
				categories:   repomock.NewMockTransactionCategoryRepository(t),
				queryService: querymock.NewMockReportQueryService(t),
			}
			if tt.setupMocks != nil {
				tt.setupMocks(m)
			}

			router := newGetCategoryDynamicReportTestRouter(report.NewService(m.accounts, m.categories, nil, m.queryService, time.UTC), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/category-dynamic"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
