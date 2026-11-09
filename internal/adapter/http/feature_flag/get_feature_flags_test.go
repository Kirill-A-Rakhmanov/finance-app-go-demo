package feature_flag_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	httpfeatureflag "finance-app-go-demo/internal/adapter/http/feature_flag"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/feature_flag"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newGetFeatureFlagsTestRouter(service *feature_flag.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpfeatureflag.NewHandler(service)
	r.GET("/feature-flag/list", h.GetFeatureFlags)
	return r
}

func TestHandler_GetFeatureFlags(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	enabledID := uuid.New()
	disabledID := uuid.New()

	type mocks struct {
		flags     *repomock.MockFeatureFlagRepository
		overrides *repomock.MockFeatureFlagOverrideRepository
	}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(m mocks)
		wantStatus int
		wantBody   string
	}{
		{
			name:    "returns the values effective for the user",
			session: session,
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return([]model.FeatureFlag{
					{ID: enabledID, Key: "registration", Enabled: true},
					{ID: disabledID, Key: "ai-receipts", Enabled: false},
				}, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, session.UserID).Return([]model.FeatureFlagOverride{
					{FeatureFlagID: disabledID, UserID: session.UserID, Enabled: true},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"featureFlags":[{"key":"registration","enabled":true},{"key":"ai-receipts","enabled":true}]}`,
		},
		{
			name:    "empty list is an array, not null",
			session: session,
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(nil, nil)
				m.overrides.EXPECT().GetByUserID(mock.Anything, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"featureFlags":[]}`,
		},
		{
			name:       "no session returns 401",
			session:    nil,
			setupMocks: func(m mocks) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "repository error returns 500",
			session: session,
			setupMocks: func(m mocks) {
				m.flags.EXPECT().GetAll(mock.Anything).Return(nil, errors.New("connection refused"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks{
				flags:     repomock.NewMockFeatureFlagRepository(t),
				overrides: repomock.NewMockFeatureFlagOverrideRepository(t),
			}
			tt.setupMocks(m)

			router := newGetFeatureFlagsTestRouter(feature_flag.NewService(m.flags, m.overrides), tt.session)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/feature-flag/list", nil))

			assert.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rec.Body.String())
			}
		})
	}
}
