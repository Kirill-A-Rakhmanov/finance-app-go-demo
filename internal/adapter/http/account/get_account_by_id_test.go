package account_test

import (
	"encoding/json"
	"fmt"
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

func newGetAccountByIDTestRouter(accountService *account.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpaccount.NewHandler(accountService)
	r.GET("/account/:accountID", h.GetAccountByID)
	return r
}

func TestHandler_GetAccountByID(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	accountID := uuid.New()

	tests := []struct {
		name       string
		session    *request.Session
		accountID  string
		setupMocks func(accountRepo *repomock.MockAccountRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:      "successful fetch",
			session:   session,
			accountID: accountID.String(),
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).
					Return(&model.Account{ID: accountID, UserID: session.UserID, Name: "My Account"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				accBody := body["account"].(map[string]any)
				assert.Equal(t, "My Account", accBody["name"])
			},
		},
		{
			name:      "no session returns 401",
			session:   nil,
			accountID: accountID.String(),
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:      "invalid accountID in path returns 400",
			session:   session,
			accountID: "not-a-uuid",
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:      "account not found maps to 404 via response.FromError",
			session:   session,
			accountID: accountID.String(),
			setupMocks: func(accountRepo *repomock.MockAccountRepository) {
				accountRepo.EXPECT().GetByIDAndUserID(mock.Anything, accountID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			tt.setupMocks(accountRepo)

			accountService := account.NewService(nil, accountRepo, nil, nil)
			router := newGetAccountByIDTestRouter(accountService, tt.session)

			path := fmt.Sprintf("/account/%s", tt.accountID)
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
