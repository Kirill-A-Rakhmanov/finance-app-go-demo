package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/request"
	httpuser "finance-app-go-demo/internal/adapter/http/user"
	"finance-app-go-demo/internal/application/user"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetCurrentUserTestRouter(userService *user.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpuser.NewHandler(userService)
	r.GET("/user/current", h.GetCurrentUser)
	return r
}

func TestHandler_GetCurrentUser(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(userRepo *repomock.MockUserRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch",
			session: session,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, session.UserID).
					Return(&model.User{ID: session.UserID, FirstName: "Jane"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				userBody := body["user"].(map[string]any)
				assert.Equal(t, "Jane", userBody["firstName"])
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "user not found maps to 404 via response.FromError",
			session: session,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				userRepo.EXPECT().GetByID(mock.Anything, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := repomock.NewMockUserRepository(t)
			tt.setupMocks(userRepo)

			userService := user.NewService(userRepo, nil, nil, nil, nil)
			router := newGetCurrentUserTestRouter(userService, tt.session)

			req := httptest.NewRequest(http.MethodGet, "/user/current", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
