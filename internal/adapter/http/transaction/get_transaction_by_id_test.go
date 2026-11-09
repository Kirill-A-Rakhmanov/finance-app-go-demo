package transaction_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransaction "finance-app-go-demo/internal/adapter/http/transaction"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newGetTransactionByIDTestRouter(transactionService *transaction.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransaction.NewHandler(transactionService)
	r.GET("/transaction/:transactionID", h.GetTransactionByID)
	return r
}

func TestHandler_GetTransactionByID(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	transactionID := uuid.New()

	tests := []struct {
		name          string
		session       *request.Session
		transactionID string
		setupMocks    func(transactionRepo *repomock.MockTransactionRepository)
		wantStatus    int
	}{
		{
			name:          "successful fetch",
			session:       session,
			transactionID: transactionID.String(),
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).
					Return(&model.Transaction{ID: transactionID}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "no session returns 401",
			session:       nil,
			transactionID: transactionID.String(),
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "invalid transactionID in path returns 400",
			session:       session,
			transactionID: "not-a-uuid",
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "transaction not found maps to 404 via response.FromError",
			session:       session,
			transactionID: transactionID.String(),
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(transactionRepo)

			transactionService := transaction.NewService(transactionRepo, nil, nil, nil, nil, nil)
			router := newGetTransactionByIDTestRouter(transactionService, tt.session)

			path := fmt.Sprintf("/transaction/%s", tt.transactionID)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
