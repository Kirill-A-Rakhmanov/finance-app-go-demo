package counterparty_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	httpcounterparty "finance-app-go-demo/internal/adapter/http/counterparty"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/counterparty"
	"finance-app-go-demo/internal/domain/model"
	repomock "finance-app-go-demo/internal/domain/repository/mock"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newGetCounterpartiesTestRouter(service *counterparty.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpcounterparty.NewHandler(service)
	r.GET("/counterparty/list", h.GetCounterparties)
	return r
}

func TestHandler_GetCounterparties(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful fetch",
			session: session,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return([]model.Counterparty{
					{ID: uuid.New(), UserID: session.UserID, Name: "Cafe"},
					{ID: uuid.New(), UserID: session.UserID, Name: "Supermarket"},
				}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string][]map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				require.Len(t, body["counterparties"], 2)
				assert.Equal(t, "Cafe", body["counterparties"][0]["name"])
			},
		},
		{
			name:    "empty list is an array, not null",
			session: session,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByUserID(mock.Anything, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				assert.JSONEq(t, `{"counterparties":[]}`, rec.Body.String())
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			router := newGetCounterpartiesTestRouter(counterparty.New(repo), tt.session)

			req := httptest.NewRequest(http.MethodGet, "/counterparty/list", nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantBody != nil {
				tt.wantBody(t, rec)
			}
		})
	}
}
