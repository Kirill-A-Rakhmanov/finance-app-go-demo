package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"finance-app-go-demo/internal/adapter/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newCORSRouter(t *testing.T, origins ...string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	m, err := middleware.NewCORSMiddleware(origins)
	require.NoError(t, err)

	r := gin.New()
	r.Use(m.Handler())
	r.GET("/ping", func(c *gin.Context) { c.Status(http.StatusOK) })

	return r
}

func corsRequest(r *gin.Engine, method, origin string, preflight bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/ping", nil)
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	if preflight {
		req.Header.Set("Access-Control-Request-Method", "GET")
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	return rec
}

func TestCORSMiddleware_AllowedOrigin(t *testing.T) {
	r := newCORSRouter(t, "http://localhost:5200", "https://app.example.com/")

	for _, origin := range []string{"http://localhost:5200", "https://app.example.com"} {
		rec := corsRequest(r, http.MethodGet, origin, false)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, origin, rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
		assert.Contains(t, rec.Header().Values("Vary"), "Origin")
	}
}

func TestCORSMiddleware_DisallowedOriginGetsNoHeaders(t *testing.T) {
	r := newCORSRouter(t, "http://localhost:5200")

	rec := corsRequest(r, http.MethodGet, "https://evil.example.com", false)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_NoOriginHeaderIsUntouched(t *testing.T) {
	r := newCORSRouter(t, "http://localhost:5200")

	rec := corsRequest(r, http.MethodGet, "", false)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, rec.Header().Values("Vary"))
}

func TestCORSMiddleware_Preflight(t *testing.T) {
	r := newCORSRouter(t, "http://localhost:5200")

	t.Run("allowed origin is answered without reaching routes", func(t *testing.T) {
		rec := corsRequest(r, http.MethodOptions, "http://localhost:5200", true)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, "http://localhost:5200", rec.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, rec.Header().Get("Access-Control-Allow-Methods"), "PATCH")
		assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.NotEmpty(t, rec.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("disallowed origin is rejected", func(t *testing.T) {
		rec := corsRequest(r, http.MethodOptions, "https://evil.example.com", true)

		assert.Equal(t, http.StatusForbidden, rec.Code)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestNewCORSMiddleware_RejectsInvalidOrigins(t *testing.T) {
	tests := []struct {
		name    string
		origins []string
	}{
		{"wildcard", []string{"*"}},
		{"no scheme", []string{"localhost:5200"}},
		{"has path", []string{"https://app.example.com/app"}},
		{"empty list", nil},
		{"only blanks", []string{" ", ""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := middleware.NewCORSMiddleware(tt.origins)

			assert.Error(t, err)
			assert.Nil(t, m)
		})
	}
}
