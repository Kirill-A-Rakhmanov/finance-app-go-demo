package transaction_test

import (
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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newDeleteTransactionTestRouter(transactionService *transaction.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransaction.NewHandler(transactionService)
	r.DELETE("/transaction/:transactionID", h.DeleteTransaction)
	return r
}

func TestHandler_DeleteTransaction(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	transactionID := uuid.New()
	accountID := uuid.New()
	transferID := uuid.New()

	tests := []struct {
		name          string
		session       *request.Session
		transactionID string
		setupMocks    func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
		)
		wantStatus int
	}{
		{
			name:          "successful deletion",
			session:       session,
			transactionID: transactionID.String(),
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, Amount: decimal.NewFromInt(30)}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, accountID).Return(decimal.NewFromInt(100), nil)
				transactionRepo.EXPECT().DeleteByID(mock.Anything, transactionID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:          "no session returns 401",
			session:       nil,
			transactionID: transactionID.String(),
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
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "deleting a transfer transaction maps to 403 via response.FromError",
			session:       session,
			transactionID: transactionID.String(),
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				transactionRepo.EXPECT().GetByIDAndUserID(mock.Anything, transactionID, session.UserID).
					Return(&model.Transaction{ID: transactionID, AccountID: accountID, Type: enum.TransactionExpense, TransferID: &transferID}, nil)
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(accountRepo, transactionRepo)

			transactionService := transaction.NewService(transactionRepo, accountRepo, nil, nil, newPassthroughTxManager(t), nil)
			router := newDeleteTransactionTestRouter(transactionService, tt.session)

			path := fmt.Sprintf("/transaction/%s", tt.transactionID)
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
