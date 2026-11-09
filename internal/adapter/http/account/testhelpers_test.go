package account_test

import (
	"finance-app-go-demo/internal/adapter/http/request"

	"github.com/gin-gonic/gin"
)

func withSession(session *request.Session) gin.HandlerFunc {
	return func(c *gin.Context) {
		if session != nil {
			ctx := request.WithSession(c.Request.Context(), *session)
			c.Request = c.Request.WithContext(ctx)
		}
		c.Next()
	}
}
