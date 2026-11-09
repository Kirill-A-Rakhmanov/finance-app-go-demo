package counterparty_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func newCreateCounterpartyTestRouter(service *counterparty.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpcounterparty.NewHandler(service)
	r.POST("/counterparty", h.CreateCounterparty)
	return r
}

func TestHandler_CreateCounterparty(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}

	tests := []struct {
		name       string
		session    *request.Session
		body       string
		setupMocks func(repo *repomock.MockCounterpartyRepository)
		wantStatus int
		wantBody   func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:    "successful creation",
			session: session,
			body:    `{"name":"Cafe"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, session.UserID, "Cafe").Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).
					Return(&model.Counterparty{ID: uuid.New(), UserID: session.UserID, Name: "Cafe"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "Cafe", body["counterparty"]["name"])
				assert.Equal(t, session.UserID.String(), body["counterparty"]["userId"])
			},
		},
		{
			name:    "no session returns 401",
			session: nil,
			body:    `{"name":"Cafe"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:    "malformed request body returns 400",
			session: session,
			body:    `{"name": }`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "empty name returns 422",
			session: session,
			body:    `{"name":""}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - validation fails before the repository is touched
			},
			wantStatus: http.StatusUnprocessableEntity,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Contains(t, body["fields"], "name")
			},
		},
		{
			name:    "name over the limit returns 422",
			session: session,
			body:    `{"name":"` + strings.Repeat("a", model.CounterpartyNameMaxLen+1) + `"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - validation fails before the repository is touched
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "duplicate name returns 409",
			session: session,
			body:    `{"name":"Cafe"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, session.UserID, "Cafe").Return(true, nil)
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			router := newCreateCounterpartyTestRouter(counterparty.New(repo), tt.session)

			req := httptest.NewRequest(http.MethodPost, "/counterparty", bytes.NewBufferString(tt.body))
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
