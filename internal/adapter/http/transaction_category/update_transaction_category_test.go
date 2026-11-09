package transaction_category_test

import (
	"bytes"
	"fmt"
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
)

func newUpdateTransactionCategoryTestRouter(service *transaction_category.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransactioncategory.NewHandler(service)
	r.PATCH("/transaction-category/:transactionCategoryID", h.UpdateTransactionCategory)
	return r
}

func TestHandler_UpdateTransactionCategory(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	categoryID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		categoryID string
		body       string
		setupMocks func(repo *repomock.MockTransactionCategoryRepository)
		wantStatus int
	}{
		{
			name:       "successful update",
			session:    session,
			categoryID: categoryID.String(),
			body:       `{"name":"Renamed"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, session.UserID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: session.UserID, Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, session.UserID, enum.Expense, "Renamed").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(&model.TransactionCategory{ID: categoryID, UserID: session.UserID, Name: "Renamed", Type: enum.Expense}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "duplicate name returns 409",
			session:    session,
			categoryID: categoryID.String(),
			body:       `{"name":"Renamed"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, session.UserID).
					Return(&model.TransactionCategory{ID: categoryID, UserID: session.UserID, Name: "Old", Type: enum.Expense}, nil)
				repo.EXPECT().ExistsByUserIDAndTypeAndName(mock.Anything, session.UserID, enum.Expense, "Renamed").Return(true, nil)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			categoryID: categoryID.String(),
			body:       `{"name":"Renamed"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid transactionCategoryID in path returns 400",
			session:    session,
			categoryID: "not-a-uuid",
			body:       `{"name":"Renamed"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed request body returns 400",
			session:    session,
			categoryID: categoryID.String(),
			body:       `{"name": }`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "category not found maps to 404 via response.FromError",
			session:    session,
			categoryID: categoryID.String(),
			body:       `{"name":"Renamed"}`,
			setupMocks: func(repo *repomock.MockTransactionCategoryRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, categoryID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockTransactionCategoryRepository(t)
			tt.setupMocks(repo)

			service := transaction_category.New(repo)
			router := newUpdateTransactionCategoryTestRouter(service, tt.session)

			path := fmt.Sprintf("/transaction-category/%s", tt.categoryID)
			req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
