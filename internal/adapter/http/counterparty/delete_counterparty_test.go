package counterparty_test

import (
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
)

func newDeleteCounterpartyTestRouter(service *counterparty.Service, session *request.Session) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(withSession(session))
	h := httpcounterparty.NewHandler(service)
	r.DELETE("/counterparty/:counterpartyID", h.DeleteCounterparty)
	return r
}

func TestHandler_DeleteCounterparty(t *testing.T) {
	session := &request.Session{UserID: uuid.New()}
	counterpartyID := uuid.New()

	tests := []struct {
		name           string
		session        *request.Session
		counterpartyID string
		setupMocks     func(repo *repomock.MockCounterpartyRepository)
		wantStatus     int
	}{
		{
			name:           "successful deletion",
			session:        session,
			counterpartyID: counterpartyID.String(),
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, session.UserID).
					Return(&model.Counterparty{ID: counterpartyID, UserID: session.UserID}, nil)
				repo.EXPECT().DeleteByID(mock.Anything, counterpartyID).Return(nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:           "no session returns 401",
			session:        nil,
			counterpartyID: counterpartyID.String(),
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - request.CurrentSession fails before the service is ever touched
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid counterpartyID in path returns 400",
			session:        session,
			counterpartyID: "not-a-uuid",
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// no calls expected - uuid.Parse fails before the service is ever touched
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:           "counterparty not found or another user's maps to 404",
			session:        session,
			counterpartyID: counterpartyID.String(),
			setupMocks: func(repo *repomock.MockCounterpartyRepository) {
				// DeleteByID must not be called
				repo.EXPECT().GetByIDAndUserID(mock.Anything, counterpartyID, session.UserID).Return(nil, nil)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := repomock.NewMockCounterpartyRepository(t)
			tt.setupMocks(repo)

			router := newDeleteCounterpartyTestRouter(counterparty.New(repo), tt.session)

			path := fmt.Sprintf("/counterparty/%s", tt.counterpartyID)
			req := httptest.NewRequest(http.MethodDelete, path, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
