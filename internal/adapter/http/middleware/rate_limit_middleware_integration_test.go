//go:build integration

package middleware_test

import (
	"context"
	"finance-app-go-demo/internal/adapter/http/middleware"
	"finance-app-go-demo/internal/testutil"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

var testClient *redis.Client

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	client, cleanup, err := testutil.StartRedis(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to start redis:", err)
		os.Exit(1)
	}
	testClient = client

	code := m.Run()

	cleanup()
	os.Exit(code)
}

func newRouter(name string, limit redis_rate.Limit) *gin.Engine {
	r := gin.New()
	r.POST("/limited", middleware.NewRateLimitMiddleware(testClient).Handler(name, limit), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return r
}

func doRequest(r *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/limited", nil)
	req.RemoteAddr = remoteAddr

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestRateLimitMiddleware_AllowsUpToLimitThenBlocks(t *testing.T) {
	r := newRouter(t.Name(), redis_rate.PerHour(3))

	for i := 0; i < 3; i++ {
		w := doRequest(r, "10.0.0.1:1234")
		require.Equal(t, http.StatusOK, w.Code, "request %d should pass", i+1)
	}

	w := doRequest(r, "10.0.0.1:1234")
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.JSONEq(t, `{"error":"too many requests, try again later"}`, w.Body.String())

	retryAfter, err := strconv.Atoi(w.Header().Get("Retry-After"))
	require.NoError(t, err, "Retry-After must be a number of seconds")
	require.Positive(t, retryAfter)
}

func TestRateLimitMiddleware_LimitsPerClientIP(t *testing.T) {
	r := newRouter(t.Name(), redis_rate.PerHour(1))

	require.Equal(t, http.StatusOK, doRequest(r, "10.0.0.1:1111").Code)
	require.Equal(t, http.StatusTooManyRequests, doRequest(r, "10.0.0.1:2222").Code, "same IP, different port shares the bucket")
	require.Equal(t, http.StatusOK, doRequest(r, "10.0.0.2:1111").Code, "another IP has its own bucket")
}

func TestRateLimitMiddleware_LimitsPerName(t *testing.T) {
	a := newRouter(t.Name()+"-a", redis_rate.PerHour(1))
	b := newRouter(t.Name()+"-b", redis_rate.PerHour(1))

	require.Equal(t, http.StatusOK, doRequest(a, "10.0.0.1:1").Code)
	require.Equal(t, http.StatusTooManyRequests, doRequest(a, "10.0.0.1:1").Code)
	require.Equal(t, http.StatusOK, doRequest(b, "10.0.0.1:1").Code, "different limiter names must not share a bucket")
}

func TestRateLimitMiddleware_FailsOpenWhenRedisIsDown(t *testing.T) {
	broken := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", MaxRetries: -1})
	defer broken.Close()

	r := gin.New()
	r.POST("/limited", middleware.NewRateLimitMiddleware(broken).Handler("down", redis_rate.PerHour(1)), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	for i := 0; i < 3; i++ {
		require.Equal(t, http.StatusOK, doRequest(r, "10.0.0.1:1").Code, "limiter errors must not block requests")
	}
}
