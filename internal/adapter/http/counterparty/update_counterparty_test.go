package counterparty_test

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func newUpdateCounterpartyTestRouter(service *counterparty.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpcounterparty.NewHandler(service)
	r.PATCH("/counterparty/:counterpartyID", h.UpdateCounterparty)
	return r
}

func TestHandler_UpdateCounterparty(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	counterpartyID := uuid.New()

	stored := func() *model.Counterparty {
		return &model.Counterparty{ID: counterpartyID, UserID: session.UserID, Name: "Old"}
	}

	tests := []struct {
		name           string
		session        *request.Session
		counterpartyID string
		body           string
		setupMocks     func(repo *repomock.MockCounterpartyRepository)
		wantStatus     int
		wantBody       func(t *testing.T, rec *httptest.ResponseRecorder)
	}{
		{
			name:           "successful update",
			session:        session,
			counterpartyID: counterpartyID.String(),
			body:           `{"name":"New"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, session.UserID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, session.UserID, "New").Return(false, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(&model.Counterparty{ID: counterpartyID, UserID: session.UserID, Name: "New"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody: func(t *testing.T, rec *httptest.ResponseRecorder) {
				var body map[string]map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
				assert.Equal(t, "New", body["counterparty"]["name"])
			},
		},
		{
			name:           "no session returns 401",
			session:        nil,
			counterpartyID: counterpartyID.String(),
			body:           `{"name":"New"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid counterpartyID in path returns 400",
			session:        session,
			counterpartyID: "not-a-uuid",
			body:           `{"name":"New"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "malformed request body returns 400",
			session:        session,
			counterpartyID: counterpartyID.String(),
			body:           `{"name": }`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - JSON binding fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "empty name returns 422",
			session:        session,
			counterpartyID: counterpartyID.String(),
			body:           `{"name":"  "}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - validation fails before the repository is touched
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "counterparty not found or another user's maps to 404",
			session:        session,
			counterpartyID: counterpartyID.String(),
			body:           `{"name":"New"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:           "duplicate name returns 409",
			session:        session,
			counterpartyID: counterpartyID.String(),
			body:           `{"name":"New"}`,
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, session.UserID).Return(stored(), nil)
				repo.EXPECT().ExistsByUserIDAndName(mock.Anything, session.UserID, "New").Return(true, nil)
			},
			wantStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			router := newUpdateCounterpartyTestRouter(counterparty.New(repo), tt.session)

			path := fmt.Sprintf("/counterparty/%s", tt.counterpartyID)
			req := httptest.NewRequest(http.MethodPatch, path, bytes.NewBufferString(tt.body))
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
