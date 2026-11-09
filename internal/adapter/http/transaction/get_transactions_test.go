package transaction_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransaction "finance-app-go-demo/internal/adapter/http/transaction"
	"finance-app-go-demo/internal/application/query"
	querymock "finance-app-go-demo/internal/application/query/mock"
	"finance-app-go-demo/internal/application/transaction"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetTransactionsTestRouter(transactionService *transaction.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransaction.NewHandler(transactionService)
	r.GET("/transaction/list", h.GetTransactions)
	return r
}

func TestHandler_GetTransactions(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()
	categoryID := uuid.New()
	otherCategoryID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		query      string
		setupMocks func(queryService *querymock.MockTransactionQueryService)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch with default limit",
			session: session,
			query:   "",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.Limit == transaction.DefaultTransactionsLimit
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, float64(transaction.DefaultTransactionsLimit), body["limit"])
			},
		},
		{
			name:    "period omitted passes nil boundaries",
			session: session,
			query:   "",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.PeriodStart == nil && c.PeriodEnd == nil
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "period in milliseconds is converted to time",
			session: session,
			query:   "?periodStart=1700000000000&periodEnd=1700086400000",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.PeriodStart != nil && c.PeriodStart.UnixMilli() == 1700000000000 &&
							c.PeriodEnd != nil && c.PeriodEnd.UnixMilli() == 1700086400000
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "only periodStart set leaves periodEnd nil",
			session: session,
			query:   "?periodStart=1700000000000",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.PeriodStart != nil && c.PeriodEnd == nil
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "valid uuid filters are parsed",
			session: session,
			query:   "?accountId=" + accountID.String() + "&transactionCategoryIds=" + categoryID.String() + "&transactionCategoryIds=" + otherCategoryID.String(),
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.AccountID != nil && *c.AccountID == accountID &&
							assert.ObjectsAreEqual([]uuid.UUID{categoryID, otherCategoryID}, c.TransactionCategoryIDs)
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "category filter omitted passes empty list",
			session: session,
			query:   "",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return len(c.TransactionCategoryIDs) == 0
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "empty category value is ignored",
			session: session,
			query:   "?transactionCategoryIds=",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return len(c.TransactionCategoryIDs) == 0
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "empty uuid filter is treated as not set",
			session: session,
			query:   "?accountId=",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().
					GetTransactions(mock.Anything, mock.MatchedBy(func(c query.GetTransactionCriteria) bool {
						return c.AccountID == nil
					})).
					Return(&query.GetTransactionsResult{Total: 0}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid accountId returns 400",
			session:    session,
			query:      "?accountId=not-a-uuid",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid transactionCategoryIds returns 400",
			session:    session,
			query:      "?transactionCategoryIds=" + categoryID.String() + "&transactionCategoryIds=123",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "no session returns 401",
			session: nil,
			query:   "",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "limit over max returns 422 with field detail",
			session: session,
			query:   "?limit=999",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				// no calls expected - Validate() short-circuits before touching any dependency
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				fields := body["fields"].(map[string]any)
				assert.Contains(t, fields, "limit")
			},
		},
		{
			name:    "unclassified query service error falls back to 500",
			session: session,
			query:   "",
			setupMocks: func(queryService *querymock.MockTransactionQueryService) {
				queryService.EXPECT().GetTransactions(mock.Anything, mock.Anything).Return(nil, errors.New("query failed"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryService := querymock.NewMockTransactionQueryService(t)
			tt.setupMocks(queryService)

			transactionService := transaction.NewService(nil, nil, nil, nil, nil, queryService)
			router := newGetTransactionsTestRouter(transactionService, tt.session)

			req := httptest.NewRequest(http.MethodGet, "/transaction/list"+tt.query, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
