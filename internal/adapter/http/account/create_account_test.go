package account_test

import (
	"bytes"
	"encoding/json"
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

func newCreateAccountTestRouter(accountService *account.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpaccount.NewHandler(accountService)
	r.POST("/account", h.CreateAccount)
	return r
}

func TestHandler_CreateAccount(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(
			accountRepo *repomock.MockAccountRepository,
			userRepo *repomock.MockUserRepository,
		)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful creation",
			session: session,
			body:    `{"name":"My Account","icon":"wallet","currency":"USD","initialBalance":100}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, session.UserID).Return(true, nil)
				accountRepo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(&model.Account{ID: uuid.New(), UserID: session.UserID, Name: "My Account"}, nil)
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
			name:    "no session returns 401",
			session: nil,
			body:    `{"name":"My Account"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"name": }`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "user not found maps to 404 via response.FromError",
			session: session,
			body:    `{"name":"My Account","currency":"USD"}`,
			setupMocks: func(
				accountRepo *repomock.MockAccountRepository,
				userRepo *repomock.MockUserRepository,
			) {
				userRepo.EXPECT().ExistsByID(mock.Anything, session.UserID).Return(false, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			accountRepo := repomock.NewMockAccountRepository(t)
			userRepo := repomock.NewMockUserRepository(t)
			tt.setupMocks(accountRepo, userRepo)

			accountService := account.NewService(nil, accountRepo, nil, userRepo)
			router := newCreateAccountTestRouter(accountService, tt.session)

			req := httptest.NewRequest(http.MethodPost, "/account", bytes.NewBufferString(tt.body))
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
