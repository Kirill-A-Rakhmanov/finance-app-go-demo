package response_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"finance-app-go-demo/internal/adapter/http/response"
	apperror "finance-app-go-demo/internal/application/error"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromError_TooManyRequests(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		wantRetryAfter string
	}{
		{"whole seconds", apperror.TooManyRequests("slow down", 90*time.Second), "90"},
		{"rounded up to a whole second", apperror.TooManyRequests("slow down", 1500*time.Millisecond), "2"},
		{"less than a second is still one second", apperror.TooManyRequests("slow down", 200*time.Millisecond), "1"},
		{"wrapped error", fmt.Errorf("login: %w", apperror.TooManyRequests("slow down", time.Minute)), "60"},
		{"no wait time means no header", apperror.TooManyRequests("slow down", 0), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()

			response.FromError(rec, tt.err)

			assert.Equal(t, http.StatusTooManyRequests, rec.Code)
			assert.Equal(t, tt.wantRetryAfter, rec.Header().Get("Retry-After"))

			var body map[string]string
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
			assert.Equal(t, "slow down", body["error"])
		})
	}
}

func TestFromError_OtherKindsHaveNoRetryAfter(t *testing.T) {
	for err, status := range map[error]int{
		apperror.Invalid("bad"):          http.StatusBadRequest,
		apperror.Conflict("exists"):      http.StatusConflict,
		apperror.Unauthorized("no"):      http.StatusUnauthorized,
		apperror.Forbidden("closed"):     http.StatusForbidden,
		apperror.NotFound("missing"):     http.StatusNotFound,
		errors.New("unexpected failure"): http.StatusInternalServerError,
	} {
		rec := httptest.NewRecorder()

		response.FromError(rec, err)

		assert.Equal(t, status, rec.Code, err.Error())
		assert.Empty(t, rec.Header().Get("Retry-After"))
	}
}

type recordingWriter struct {
	*httptest.ResponseRecorder
	recorded error
}

func (w *recordingWriter) RecordError(err error) {
	w.recorded = err
}

func TestInternal_HidesTheRealErrorFromTheClient(t *testing.T) {
	rec := httptest.NewRecorder()

	response.Internal(rec, errors.New(`pq: duplicate key value violates unique constraint "uni_users_email"`))

	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal server error", body["error"])
	assert.NotContains(t, rec.Body.String(), "uni_users_email")
}

func TestInternal_PassesTheRealErrorToTheWriterThatRecordsIt(t *testing.T) {
	w := &recordingWriter{ResponseRecorder: httptest.NewRecorder()}
	cause := errors.New("connection refused")

	response.FromError(w, fmt.Errorf("load accounts: %w", cause))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.ErrorIs(t, w.recorded, cause)
	assert.NotContains(t, w.Body.String(), "connection refused")
}

func TestFromError_OnlyUnexpectedErrorsAreRecorded(t *testing.T) {
	w := &recordingWriter{ResponseRecorder: httptest.NewRecorder()}

	response.FromError(w, apperror.Conflict("exists"))

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.NoError(t, w.recorded)
}
