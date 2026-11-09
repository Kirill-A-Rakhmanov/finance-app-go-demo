package counterparty_test

import (
	"finance-app-go-demo/internal/adapter/http/request"

	"github.com/gin-gonic/gin"
)

// withSession injects the given session into the request context, standing in
// for the real authMiddleware so handlers can be exercised without the rest of
// the middleware stack. A nil session leaves the context untouched, simulating
// an unauthenticated request.
func withSession(session *request.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		if session != nil {
			ctx := request.WithSession(c.Request.Context(), *session)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
