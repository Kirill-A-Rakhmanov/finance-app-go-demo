package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

type RateLimitMiddleware struct {
	limiter *redis_rate.Limiter
}

func NewRateLimitMiddleware(redisClient *redis.Client) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter: redis_rate.NewLimiter(redisClient),
	}
}

func (m *RateLimitMiddleware) Handler(name string, limit redis_rate.Limit) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := fmt.Sprintf("ratelimit:%s:%s", name, c.ClientIP())

		res, err := m.limiter.Allow(c.Request.Context(), key, limit)
		if err != nil {
			c.Next()
			return
		}

		if res.Allowed == 0 {
			c.Header("Retry-After", fmt.Sprintf("%.0f", res.RetryAfter.Seconds()))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests, try again later",
			})
			return
		}

		c.Next()
	}
}
