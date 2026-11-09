package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

const corsMaxAge = "43200" // 12h, the longest browsers honour (Chromium caps it lower)

type CORSMiddleware struct {
	allowedOrigins map[string]struct{}
}

func NewCORSMiddleware(allowedOrigins []string) (*CORSMiddleware, error) {
	origins := make(map[string]struct{}, len(allowedOrigins))

	for _, raw := range allowedOrigins {
		origin := strings.TrimRight(strings.TrimSpace(raw), "/")
		if origin == "" {
			continue
		}

		u, err := url.Parse(origin)
		if err != nil || u.Scheme == "" || u.Host == "" || u.Path != "" {
			return nil, fmt.Errorf("invalid CORS origin %q, expected scheme://host[:port] without wildcard or path", raw)
		}

		origins[origin] = struct{}{}
	}

	if len(origins) == 0 {
		return nil, fmt.Errorf("at least one CORS origin is required")
	}

	return &CORSMiddleware{allowedOrigins: origins}, nil
}

func (m *CORSMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			c.Next()
			return
		}

		c.Writer.Header().Add("Vary", "Origin")

		if _, ok := m.allowedOrigins[origin]; !ok {
			if isPreflight(c.Request) {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}

			c.Next()
			return
		}

		h := c.Writer.Header()
		h.Set("Access-Control-Allow-Origin", origin)
		h.Set("Access-Control-Allow-Credentials", "true")

		if isPreflight(c.Request) {
			h.Add("Vary", "Access-Control-Request-Method")
			h.Add("Vary", "Access-Control-Request-Headers")
			h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			h.Set("Access-Control-Allow-Headers", "Content-Type")
			h.Set("Access-Control-Max-Age", corsMaxAge)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}
