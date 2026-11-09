package middleware

import (
	"finance-app-go-demo/internal/adapter/http/cookie"
	"finance-app-go-demo/internal/adapter/http/request"
	"finance-app-go-demo/internal/application/auth"
	"finance-app-go-demo/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SessionData struct {
	UserID    uuid.UUID `json:"userId"`
	SessionID string    `json:"sessionId"`
}

type AuthMiddleware struct {
	authService *auth.Service
}

func NewAuthMiddleware(authService *auth.Service) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

func (m *AuthMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		acc, err := c.Cookie(cookie.AccessTokenName)

		if err != nil || acc == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "no access token",
			})
			return
		}

		result, err := m.authService.VerifyAuth(c.Request.Context(), auth.VerifyInput{
			AccessToken: acc,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		session := request.Session{
			UserID:    result.UserID,
			SessionID: result.SessionID,
		}

		ctx := request.WithSession(c.Request.Context(), session)
		ctx = logger.WithFields(ctx, "userID", session.UserID, "sessionID", session.SessionID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
