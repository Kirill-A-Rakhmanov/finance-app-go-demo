package middleware

import (
	"bytes"
	"encoding/json"
	"finance-app-go-demo/pkg/logger"
	"fmt"
	"net/http"
	"runtime/debug"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

type LoggingMiddleware struct {
	log        logger.Logger
	quietPaths []string
}

func NewLoggingMiddleware(log logger.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		log:        log,
		quietPaths: []string{"/health"},
	}
}

type bodyCapturingWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
	err  error
}

func (w *bodyCapturingWriter) RecordError(err error) {
	w.err = err
}

func (w *bodyCapturingWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func errorMessageFromBody(body []byte) string {
	var payload struct {
		Error string `json:"error"`
	}

	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	return payload.Error
}

func (m *LoggingMiddleware) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		bw := &bodyCapturingWriter{ResponseWriter: c.Writer, body: &bytes.Buffer{}}
		c.Writer = bw

		c.Next()

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		status := c.Writer.Status()
		ctx := c.Request.Context()

		if status < http.StatusBadRequest && slices.Contains(m.quietPaths, path) {
			return
		}

		fields := []any{
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"duration", time.Since(start).String(),
			"clientIP", c.ClientIP(),
		}

		if status >= 400 {
			if bw.err != nil {
				fields = append(fields, "error", bw.err.Error())
			} else if msg := errorMessageFromBody(bw.body.Bytes()); msg != "" {
				fields = append(fields, "error", msg)
			}
		}

		switch {
		case status >= 500:
			m.log.ErrorContext(ctx, "http request", fields...)
		case status >= 400:
			m.log.WarnContext(ctx, "http request", fields...)
		default:
			m.log.InfoContext(ctx, "http request", fields...)
		}
	}
}

func (m *LoggingMiddleware) Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}

			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			m.log.ErrorContext(c.Request.Context(), "panic recovered",
				"panic", fmt.Sprint(rec),
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
				"stack", string(debug.Stack()),
			)

			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}()

		c.Next()
	}
}
