package transaction_test

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransaction "finance-app-go-demo/internal/adapter/http/transaction"
	"finance-app-go-demo/internal/application/transaction"
	"finance-app-go-demo/internal/domain/enum"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newUpdateTransactionTestRouter(transactionService *transaction.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransaction.NewHandler(transactionService)
	r.PATCH("/transaction/:transactionID", h.UpdateTransaction)
	return r
}

func TestHandler_UpdateTransaction(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	transactionID := uuid.New()
	accountID := uuid.New()

	validBody := fmt.Sprintf(`{"accountId":"%s","amount":50,"type":"%s"}`, accountID, enum.TransactionExpense)

	tests := []struct {
		name          string
		session       *request.Session
		transactionID string
		body          string
		setupMocks    func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
		)
		wantStatus int
	}{
		{
			name:          "successful update",
			session:       session,
			transactionID: transactionID.String(),
			body:          validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID, IsNegativeBalanceEnabled: true}, nil)
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense}, nil)
				transactionRepo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "no session returns 401",
			session:       nil,
			transactionID: transactionID.String(),
			body:          validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:          "invalid transactionID in path returns 400",
			session:       session,
			transactionID: "not-a-uuid",
			body:          validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "malformed request body returns 400",
			session:       session,
			transactionID: transactionID.String(),
			body:          `{"amount": }`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "transaction not found maps to 404 via response.FromError",
			session:       session,
			transactionID: transactionID.String(),
			body:          validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID}, nil)
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(accountRepo, transactionRepo)

			transactionService := transaction.NewService(transactionRepo, accountRepo, nil, nil, newPassthroughTxManager(t), nil)
			router := newUpdateTransactionTestRouter(transactionService, tt.session)

			path := fmt.Sprintf("/transaction/%s", tt.transactionID)
			req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
