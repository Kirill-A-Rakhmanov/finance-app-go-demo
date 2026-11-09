package transaction_category_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransactioncategory "finance-app-go-demo/internal/adapter/http/transaction_category"
	"finance-app-go-demo/internal/application/transaction_category"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetTransactionCategoriesTestRouter(service *transaction_category.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransactioncategory.NewHandler(service)
	r.GET("/transaction-category/list", h.GetTransactionCategories)
	return r
}

func TestHandler_GetTransactionCategories(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch",
			session: session,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return([]model.TransactionCategory{
					{ID: uuid.New(), UserID: session.UserID, Name: "Groceries"},
					{ID: uuid.New(), UserID: session.UserID, Name: "Salary"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				categories := body["transactionCategories"].([]any)
				assert.Len(t, categories, 2)
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "unclassified repository error falls back to 500",
			session: session,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return(nil, errors.New("connection refused"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)
			router := newGetTransactionCategoriesTestRouter(service, tt.session)

			req := httptest.NewRequest(http.MethodGet, "/transaction-category/list", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
