package transfer_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransfer "finance-app-go-demo/internal/adapter/http/transfer"
	"finance-app-go-demo/internal/application/transfer"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCreateTransferTestRouter(transferService *transfer.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransfer.NewHandler(transferService)
	r.POST("/transfer", h.CreateTransfer)
	return r
}

func TestHandler_CreateTransfer(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	fromAccountID := uuid.New()
	toAccountID := uuid.New()
	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	validBody := `{"fromAccountId":"` + fromAccountID.String() + `","toAccountId":"` + toAccountID.String() + `","fromAmount":50,"toAmount":50}`

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
			transferRepo *repomock.MockTransferRepository,
			txManager *transactionmock.MockManager,
		)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful transfer",
			session: session,
			body:    validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, session.UserID).
					Return(&model.Account{ID: fromAccountID, UserID: session.UserID, Currency: usd}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, session.UserID).
					Return(&model.Account{ID: toAccountID, UserID: session.UserID, Currency: usd}, nil)
				transactionRepo.EXPECT().GetBalanceByAccountID(mock.Anything, fromAccountID).Return(decimal.NewFromInt(100), nil)
				transferRepo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(&model.Transfer{ID: uuid.New(), FromAccountID: fromAccountID, ToAccountID: toAccountID}, nil)
				transactionRepo.EXPECT().BulkCreate(mock.Anything, mock.Anything).
					Return([]model.Transaction{{}, {}}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				transactions := body["transactions"].([]any)
				assert.Len(t, transactions, 2)
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
				txManager *transactionmock.MockManager,
			) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"fromAmount": }`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
				txManager *transactionmock.MockManager,
			) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "transfer to same account maps to 400 via response.FromError",
			session: session,
			body:    `{"fromAccountId":"` + fromAccountID.String() + `","toAccountId":"` + fromAccountID.String() + `","fromAmount":50,"toAmount":50}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
				txManager *transactionmock.MockManager,
			) {
				// no calls expected - same-account check happens before entering the transaction
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "from account not found maps to 404 via response.FromError",
			session: session,
			body:    validBody,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, fromAccountID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			transactionRepo := repomock.NewMockTransactionRepository(t)
			transferRepo := repomock.NewMockTransferRepository(t)
			txManager := transactionmock.NewMockManager(t)
			tt.setupMocks(accountRepo, transactionRepo, transferRepo, txManager)

			transferService := transfer.NewService(transferRepo, transactionRepo, accountRepo, txManager)
			router := newCreateTransferTestRouter(transferService, tt.session)

			req := httptest.NewRequest(http.MethodPost, "/transfer", bytes.NewBufferString(tt.body))
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
