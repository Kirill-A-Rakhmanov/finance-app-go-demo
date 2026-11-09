package middleware

import (
	"regexp"

	"finance-app-go-demo/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

var validRequestID = regexp.MustCompile(`^[A-Za-z0-9._-]{8,64}$`)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(RequestIDHeader)
		if !validRequestID.MatchString(id) {
			id = uuid.NewString()
		}

		c.Header(RequestIDHeader, id)

		ctx := logger.WithFields(c.Request.Context(), "requestID", id)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
