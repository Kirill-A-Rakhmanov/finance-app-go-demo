package transaction_category_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransactioncategory "finance-app-go-demo/internal/adapter/http/transaction_category"
	"finance-app-go-demo/internal/application/transaction_category"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCreateTransactionCategoryTestRouter(service *transaction_category.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransactioncategory.NewHandler(service)
	r.POST("/transaction-category", h.CreateTransactionCategory)
	return r
}

func TestHandler_CreateTransactionCategory(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful creation",
			session: session,
			body:    `{"name":"Groceries","type":"expense"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, session.UserID, enum.Expense, "Groceries").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(&model.TransactionCategory{ID: uuid.New(), UserID: session.UserID, Name: "Groceries"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				catBody := body["transactionCategory"].(map[string]any)
				assert.Equal(t, "Groceries", catBody["name"])
			},
		},
		{
			name:    "duplicate name returns 409",
			session: session,
			body:    `{"name":"Groceries","type":"expense"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, session.UserID, enum.Expense, "Groceries").Return(true, nil)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    `{"name":"Groceries","type":"expense"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"name": }`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "validation error returns 422 with field detail",
			session: session,
			body:    `{"name":"","type":"unknown"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - Validate() short-circuits before touching any dependency
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				fields := body["fields"].(map[string]any)
				assert.Contains(t, fields, "name")
				assert.Contains(t, fields, "type")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)
			router := newCreateTransactionCategoryTestRouter(service, tt.session)

			req := httptest.NewRequest(http.MethodPost, "/transaction-category", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
