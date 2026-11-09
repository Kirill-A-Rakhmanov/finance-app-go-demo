package user_test

import (
	"bytes"
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

func newUpdateUserTestRouter(userService *user.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpuser.NewHandler(userService)
	r.PATCH("/user/current", h.UpdateUser)
	return r
}

func TestHandler_UpdateUser(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(userRepo *repomock.MockUserRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful update",
			session: session,
			body:    `{"firstName":"Jane","lastName":"Smith"}`,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
				existing := &model.User{ID: session.UserID, FirstName: "John", LastName: "Doe"}
				userRepo.EXPECT().GetByID(mock.Anything, session.UserID).Return(existing, nil)
				userRepo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(&model.User{ID: session.UserID, FirstName: "Jane", LastName: "Smith"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				userBody := body["user"].(map[string]any)
				assert.Equal(t, "Jane", userBody["firstName"])
				assert.Equal(t, "Smith", userBody["lastName"])
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    `{"firstName":"Jane","lastName":"Smith"}`,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"firstName": }`,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "validation error returns 422",
			session: session,
			body:    `{"firstName":"","lastName":""}`,
			setupMocks: func(userRepo *repomock.MockUserRepository) {
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "user not found maps to 404",
			session: session,
			body:    `{"firstName":"Jane","lastName":"Smith"}`,
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
			router := newUpdateUserTestRouter(userService, tt.session)

			req := httptest.NewRequest(http.MethodPatch, "/user/current", bytes.NewBufferString(tt.body))
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
