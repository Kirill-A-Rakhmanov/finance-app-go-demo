package transaction_test

import (
	"bytes"
	"encoding/json"
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
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCreateExpenseTransactionTestRouter(transactionService *transaction.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransaction.NewHandler(transactionService)
	r.POST("/transaction/expense", h.CreateExpenseTransaction)
	return r
}

func TestHandler_CreateExpenseTransaction(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
		)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful creation",
			session: session,
			body:    `{"accountId":"` + accountID.String() + `","amount":50}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID, IsNegativeBalanceEnabled: true}, nil)
				transactionRepo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(&model.Transaction{ID: uuid.New(), AccountID: accountID, Amount: decimal.NewFromInt(50)}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				txBody := body["transaction"].(map[string]any)
				assert.Equal(t, float64(50), txBody["amount"])
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    `{"accountId":"` + accountID.String() + `","amount":50}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"amount": }`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "account not found maps to 404 via response.FromError",
			session: session,
			body:    `{"accountId":"` + accountID.String() + `","amount":50}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
			) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).Return(nil, nil)
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
			router := newCreateExpenseTransactionTestRouter(transactionService, tt.session)

			req := httptest.NewRequest(http.MethodPost, "/transaction/expense", bytes.NewBufferString(tt.body))
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
