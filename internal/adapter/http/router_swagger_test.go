package http_test

import (
	"io"
	"log/slog"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	httprouter "finance-app-go-demo/internal/adapter/http"
	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRouter_SwaggerIsServedOnlyWhenEnabled(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name    string
		enabled bool
		want    int
	}{
		{"enabled", true, nethttp.StatusOK},
		{"disabled", false, nethttp.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cors, err := middleware.NewCORSMiddleware([]string{"http://localhost:5200"})
			require.NoError(t, err)

			router, err := httprouter.NewRouter(
				&middleware.AuthMiddleware{}, middleware.NewRateLimitMiddleware(unreachableRedis()), middleware.NewLoggingMiddleware(logger.New(slog.New(slog.NewTextHandler(io.Discard, nil)))), cors,
				nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil,
				nil,
				redis_rate.PerMinute(300),
				tt.enabled,
			)
			require.NoError(t, err)

			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(nethttp.MethodGet, "/api/swagger/index.html", nil))

			assert.Equal(t, tt.want, rec.Code)
		})
	}
}

// unreachableRedis makes the rate limiter fail open immediately, so router tests
// do not need a Redis instance.
func unreachableRedis() *redis.Client {
	return redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1, DialTimeout: 50 * time.Millisecond})
}
