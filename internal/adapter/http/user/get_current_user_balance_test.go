package user_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httpuser "finance-app-go-demo/internal/adapter/http/user"
	"finance-app-go-demo/internal/application/user"
	repomock "finance-app-go-demo/internal/domain/repository/mock"
	"finance-app-go-demo/internal/domain/valueObject"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetCurrentUserBalanceTestRouter(userService *user.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpuser.NewHandler(userService)
	r.GET("/user/current/balance", h.GetCurrentUserBalance)
	return r
}

func TestHandler_GetCurrentUserBalance(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	usd, err := valueObject.NewCurrency(valueObject.CurrencyUSD)
	require.NoError(t, err)

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(transactionRepo *repomock.MockTransactionRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch",
			session: session,
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetBalanceByUserID(mock.Anything, session.UserID).Return([]valueObject.Balance{
					{Currency: usd, Amount: decimal.NewFromInt(150)},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				balances := body["balances"].([]any)
				require.Len(t, balances, 1)
				balance := balances[0].(map[string]any)
				assert.Equal(t, float64(150), balance["amount"])
				currency := balance["currency"].(map[string]any)
				assert.Equal(t, "USD", currency["code"])
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "unclassified repository error falls back to 500",
			session: session,
			setupMocks: func(transactionRepo *repomock.MockTransactionRepository) {
				transactionRepo.EXPECT().GetBalanceByUserID(mock.Anything, session.UserID).Return(nil, errors.New("connection refused"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transactionRepo := repomock.NewMockTransactionRepository(t)
			tt.setupMocks(transactionRepo)

			userService := user.NewService(nil, transactionRepo, nil, nil, nil)
			router := newGetCurrentUserBalanceTestRouter(userService, tt.session)

			req := httptest.NewRequest(http.MethodGet, "/user/current/balance", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
