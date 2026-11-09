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

func newGetCategoryReportTestRouter(service *report.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpreport.NewHandler(service)
	r.GET("/report/categories", h.GetCategoryReport)
	return r
}

func TestHandler_GetCategoryReport(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	categoryID := uuid.New()
	rentID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		query      string
		setupMocks func(accounts *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful report",
			session: session,
			query:   "?type=expense&currency=USD",
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.Anything).
					Return(&query.GetCategoryReportResult{
						Items: []query.CategoryReportItem{
							{TransactionCategory: &model.TransactionCategory{ID: rentID, Type: enum.Expense}, Amount: decimal.RequireFromString("600")},
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
				assert.Equal(t, rentID.String(), item["transactionCategoryId"])
				assert.Equal(t, float64(600), item["amount"])
				assert.Equal(t, float64(60), item["share"])

				uncategorized := items[1].(map[string]any)
				assert.Contains(t, uncategorized, "transactionCategoryId", "the key is present with a null value")
				assert.Nil(t, uncategorized["transactionCategoryId"])
				assert.Equal(t, float64(400), uncategorized["amount"])
				assert.Equal(t, float64(40), uncategorized["share"])

				categories := body["transactionCategories"].([]any)
				require.Len(t, categories, 1, "only real categories are listed")
				assert.Equal(t, rentID.String(), categories[0].(map[string]any)["id"])
			},
		},
		{
			name:    "empty report returns an empty array",
			session: session,
			query:   "?type=income&currency=USD",
			setupMocks: func(_ *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				queryService.EXPECT().GetCategoryReport(mock.Anything, mock.Anything).Return(&query.GetCategoryReportResult{}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, []any{}, body["items"])
				assert.Equal(t, []any{}, body["transactionCategories"])
				assert.Equal(t, float64(0), body["total"])
			},
		},
		{
			name:    "query params are converted into criteria",
			session: session,
			query: "?type=income&currency=USD&groupSubcategories=true&periodStart=1700000000000&periodEnd=1700086400000" +
				"&accountIds=" + accountID.String() + "&transactionCategoryIds=" + categoryID.String(),
			setupMocks: func(accounts *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, queryService *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyUSD}}}, nil)
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, session.UserID).
					Return([]model.TransactionCategory{{ID: categoryID}}, nil)
				queryService.EXPECT().
					GetCategoryReport(mock.Anything, mock.MatchedBy(func(c query.GetCategoryReportCriteria) bool {
						return c.UserID == session.UserID &&
							c.TransactionType == enum.TransactionIncome &&
							c.Currency == valueObject.CurrencyUSD &&
							c.GroupSubcategories &&
							c.PeriodStart != nil && c.PeriodStart.UnixMilli() == 1700000000000 &&
							c.PeriodEnd != nil && c.PeriodEnd.UnixMilli() == 1700086400000 &&
							assert.ObjectsAreEqual([]uuid.UUID{accountID}, c.AccountIDs) &&
							assert.ObjectsAreEqual([]uuid.UUID{categoryID}, c.TransactionCategoryIDs)
					})).
					Return(&query.GetCategoryReportResult{}, nil)
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
			name:       "invalid category id returns 400",
			session:    session,
			query:      "?type=expense&currency=USD&transactionCategoryIds=not-a-uuid",
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
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).Return([]model.Account{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "foreign category returns 404",
			session: session,
			query:   "?type=expense&currency=USD&transactionCategoryIds=" + categoryID.String(),
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, session.UserID).Return([]model.TransactionCategory{}, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:    "account in another currency returns 400",
			session: session,
			query:   "?type=expense&currency=USD&accountIds=" + accountID.String(),
			setupMocks: func(accounts *repomock.MockAccountRepository, _ *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				accounts.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{accountID}, session.UserID).
					Return([]model.Account{{ID: accountID, Currency: valueObject.Currency{Code: valueObject.CurrencyEUR}}}, nil)
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "subcategory filter with grouping returns 400",
			session: session,
			query:   "?type=expense&currency=USD&groupSubcategories=true&transactionCategoryIds=" + categoryID.String(),
			setupMocks: func(_ *repomock.MockAccountRepository, categories *repomock.MockTransactionCategoryRepository, _ *querymock.MockReportQueryService) {
				parentID := uuid.New()
				categories.EXPECT().GetByIDsAndUserID(mock.Anything, []uuid.UUID{categoryID}, session.UserID).
					Return([]model.TransactionCategory{{ID: categoryID, ParentID: &parentID}}, nil)
			},
			wantStatus: http.StatusBadRequest,
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

			router := newGetCategoryReportTestRouter(report.NewService(accounts, categories, nil, queryService, time.UTC), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/report/categories"+tt.query, nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
