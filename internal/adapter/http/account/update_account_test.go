package account_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	httpaccount "finance-app-go-demo/internal/adapter/http/account"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/account"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	transactionmock "finance-app-go-demo/pkg/transaction/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func runTx(txManager *transactionmock.MockManager) {
	txManager.EXPECT().
		Wrap(mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, txFunc func(context.Context) error) error {
			return txFunc(ctx)
		})
}

func newUpdateAccountTestRouter(accountService *account.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpaccount.NewHandler(accountService)
	r.PATCH("/account/:accountID", h.UpdateAccount)
	return r
}

func TestHandler_UpdateAccount(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		accountID  string
		body       string
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			txManager *transactionmock.MockManager,
		)
		wantStatus int
	}{
		{
			name:      "successful update",
			session:   session,
			accountID: accountID.String(),
			body:      `{"name":"Renamed"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID}, nil)
				accountRepo.EXPECT().Update(mock.Anything, mock.Anything).
					RunAndReturn(func(ctx context.Context, acc *model.Account) (*model.Account, error) {
						return acc, nil
					})
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "no session returns 401",
			session:   nil,
			accountID: accountID.String(),
			body:      `{"name":"Renamed"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "invalid accountID in path returns 400",
			session:   session,
			accountID: "not-a-uuid",
			body:      `{"name":"Renamed"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "malformed request body returns 400",
			session:   session,
			accountID: accountID.String(),
			body:      `{"name": }`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "account not found maps to 404 via response.FromError",
			session:   session,
			accountID: accountID.String(),
			body:      `{"name":"Renamed"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				txManager *transactionmock.MockManager,
			) {
				runTx(txManager)
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			txManager := transactionmock.NewMockManager(t)
			tt.setupMocks(accountRepo, txManager)

			accountService := account.NewService(txManager, accountRepo, nil, nil)
			router := newUpdateAccountTestRouter(accountService, tt.session)

			path := fmt.Sprintf("/account/%s", tt.accountID)
			req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
