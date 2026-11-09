package health

import (
	"context"
	"finance-app-go-demo/ent"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const checkTimeout = 2 * time.Second

type checkFunc func(ctx context.Context) error

type Handler struct {
	checks  map[string]checkFunc
	timeout time.Duration
}

func NewHandler(dbClient *ent.Client, redisClient *redis.Client) *Handler {
	return &Handler{
		checks: map[string]checkFunc{
			"postgres": func(ctx context.Context) error {
				_, err := dbClient.User.Query().Limit(1).Exist(ctx)
				return err
			},
			"redis": func(ctx context.Context) error {
				return redisClient.Ping(ctx).Err()
			},
		},
		timeout: checkTimeout,
	}
}

type CheckResult struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type HealthResponse struct {
	Status string                 `json:"status"`
	Checks map[string]CheckResult `json:"checks"`
}

// @Summary Health check
// @Description Check application and dependency (Postgres, Redis) health
// @Tags health
// @Produce json
// @Success 200 {object} HealthResponse
// @Failure 503 {object} HealthResponse
// @Router /health [get]
func (h *Handler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.timeout)
	defer cancel()

	var (
		mu      sync.Mutex
		wg      sync.WaitGroup
		checks  = make(map[string]CheckResult, len(h.checks))
		healthy = true
	)

	for name, check := range h.checks {
		wg.Add(1)
		go func() {
			defer wg.Done()

			err := check(ctx)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				checks[name] = CheckResult{Status: "down", Error: err.Error()}
				healthy = false
				return
			}

			checks[name] = CheckResult{Status: "up"}
		}()
	}

	wg.Wait()

	status := http.StatusOK
	overall := "ok"
	if !healthy {
		status = http.StatusServiceUnavailable
		overall = "unavailable"
	}

	c.JSON(status, HealthResponse{Status: overall, Checks: checks})
}
