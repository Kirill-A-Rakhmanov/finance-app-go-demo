package health

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func serveHealth(t *testing.T, checks map[string]checkFunc, timeout time.Duration) (*httptest.ResponseRecorder, HealthResponse, time.Duration) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	h := &Handler{checks: checks, timeout: timeout}
	r.GET("/health", h.Health)

	rec := httptest.NewRecorder()
	start := time.Now()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	elapsed := time.Since(start)

	var body HealthResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	return rec, body, elapsed
}

func up(context.Context) error { return nil }

func TestHandler_Health(t *testing.T) {
	t.Run("all dependencies up", func(t *testing.T) {
		rec, body, _ := serveHealth(t, map[string]checkFunc{"postgres": up, "redis": up}, time.Second)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "ok", body.Status)
		assert.Equal(t, "up", body.Checks["postgres"].Status)
		assert.Equal(t, "up", body.Checks["redis"].Status)
	})

	t.Run("a failing dependency makes the check unavailable", func(t *testing.T) {
		down := func(context.Context) error { return errors.New("connection refused") }

		rec, body, _ := serveHealth(t, map[string]checkFunc{"postgres": down, "redis": up}, time.Second)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Equal(t, "unavailable", body.Status)
		assert.Equal(t, "down", body.Checks["postgres"].Status)
		assert.Equal(t, "connection refused", body.Checks["postgres"].Error)
		assert.Equal(t, "up", body.Checks["redis"].Status)
	})

	t.Run("a hung dependency is cut off by the timeout", func(t *testing.T) {
		hung := func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		}

		rec, body, elapsed := serveHealth(t, map[string]checkFunc{"postgres": hung, "redis": up}, 50*time.Millisecond)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Equal(t, "down", body.Checks["postgres"].Status)
		assert.Equal(t, "up", body.Checks["redis"].Status)
		assert.Less(t, elapsed, time.Second)
	})
}
