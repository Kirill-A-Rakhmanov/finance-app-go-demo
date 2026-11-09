package account_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpaccount "finance-app-go-demo/internal/adapter/http/account"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/account"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetAccountsByUserIDTestRouter(accountService *account.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpaccount.NewHandler(accountService)
	r.GET("/account/list", h.GetAccountsByUserID)
	return r
}

func TestHandler_GetAccountsByUserID(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(accountRepo *repomock.MockAccountRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch",
			session: session,
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return([]model.Account{
					{ID: uuid.New(), UserID: session.UserID, Name: "Cash"},
					{ID: uuid.New(), UserID: session.UserID, Name: "Card"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				accounts := body["accounts"].([]any)
				assert.Len(t, accounts, 2)
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "unclassified repository error falls back to 500",
			session: session,
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return(nil, errors.New("connection refused"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			tt.setupMocks(accountRepo)

			accountService := account.NewService(nil, accountRepo, nil, nil)
			router := newGetAccountsByUserIDTestRouter(accountService, tt.session)

			req := httptest.NewRequest(http.MethodGet, "/account/list", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
