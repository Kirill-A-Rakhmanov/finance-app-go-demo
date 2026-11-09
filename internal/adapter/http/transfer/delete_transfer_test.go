package transfer_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httptransfer "finance-app-go-demo/internal/adapter/http/transfer"
	"finance-app-go-demo/internal/application/transfer"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newDeleteTransferTestRouter(transferService *transfer.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httptransfer.NewHandler(transferService)
	r.DELETE("/transfer/:transferID", h.DeleteTransfer)
	return r
}

func TestHandler_DeleteTransfer(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	transferID := uuid.New()
	toAccountID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		transferID string
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			transactionRepo *repomock.MockTransactionRepository,
			transferRepo *repomock.MockTransferRepository,
		)
		wantStatus int
	}{
		{
			name:       "successful deletion",
			session:    session,
			transferID: transferID.String(),
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).
					Return(&model.Transfer{ID: transferID, ToAccountID: toAccountID}, nil)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, toAccountID, session.UserID).
					Return(&model.Account{ID: toAccountID, UserID: session.UserID, IsNegativeBalanceEnabled: true}, nil)
				transferRepo.EXPECT().DeleteByID(mock.Anything, transferID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			transferID: transferID.String(),
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid transferID in path returns 400",
			session:    session,
			transferID: "not-a-uuid",
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "transfer not found maps to 404 via response.FromError",
			session:    session,
			transferID: transferID.String(),
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				transactionRepo *repomock.MockTransactionRepository,
				transferRepo *repomock.MockTransferRepository,
			) {
				transferRepo.EXPECT().GetByID(mock.Anything, transferID).Return(nil, nil)
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
			txManager.EXPECT().
				Wrap(mock.Anything, mock.Anything).
				RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
					return txFunc(ctx)
				}).
				Maybe()
			txManager.EXPECT().Lock(mock.Anything, mock.Anything).Return(nil).Maybe()
			tt.setupMocks(accountRepo, transactionRepo, transferRepo)

			transferService := transfer.NewService(transferRepo, transactionRepo, accountRepo, txManager)
			router := newDeleteTransferTestRouter(transferService, tt.session)

			path := fmt.Sprintf("/transfer/%s", tt.transferID)
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
