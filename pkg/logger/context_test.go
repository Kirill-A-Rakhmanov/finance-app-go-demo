package logger_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"finance-app-go-demo/pkg/logger"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) logger.Logger {
	handler := logger.NewContextHandler(slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	return logger.New(slog.New(handler))
}

func decodeLine(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()

	var line map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &line))

	return line
}

func TestContextHandler_AddsFieldsFromContext(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	ctx := logger.WithFields(context.Background(), "requestID", "abc123", "attempt", 2)
	log.InfoContext(ctx, "hello", "extra", "x")

	line := decodeLine(t, &buf)
	assert.Equal(t, "hello", line["msg"])
	assert.Equal(t, "abc123", line["requestID"])
	assert.EqualValues(t, 2, line["attempt"])
	assert.Equal(t, "x", line["extra"])
}

func TestContextHandler_FieldsAccumulateAcrossCalls(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	base := logger.WithFields(context.Background(), "requestID", "abc123")
	withUser := logger.WithFields(base, "userID", "u-1")

	log.WarnContext(withUser, "both")
	both := decodeLine(t, &buf)
	assert.Equal(t, "abc123", both["requestID"])
	assert.Equal(t, "u-1", both["userID"])

	buf.Reset()
	log.WarnContext(base, "parent untouched")
	parent := decodeLine(t, &buf)
	assert.Equal(t, "abc123", parent["requestID"])
	assert.NotContains(t, parent, "userID")
}

func TestContextHandler_AllContextMethodsCarryFields(t *testing.T) {
	ctx := logger.WithFields(context.Background(), "requestID", "abc123")

	calls := map[string]func(log logger.Logger){
		"debug": func(log logger.Logger) { log.DebugContext(ctx, "m") },
		"info":  func(log logger.Logger) { log.InfoContext(ctx, "m") },
		"warn":  func(log logger.Logger) { log.WarnContext(ctx, "m") },
		"error": func(log logger.Logger) { log.ErrorContext(ctx, "m") },
	}

	for level, call := range calls {
		t.Run(level, func(t *testing.T) {
			var buf bytes.Buffer
			call(newTestLogger(&buf))

			line := decodeLine(t, &buf)
			assert.Equal(t, "abc123", line["requestID"])
			assert.Equal(t, level, lowerLevel(line["level"]))
		})
	}
}

func lowerLevel(level any) string {
	switch level {
	case "DEBUG":
		return "debug"
	case "INFO":
		return "info"
	case "WARN":
		return "warn"
	case "ERROR":
		return "error"
	}

	return ""
}

func TestContextHandler_PlainMethodsAndEmptyContextAddNothing(t *testing.T) {
	var buf bytes.Buffer
	log := newTestLogger(&buf)

	log.Info("startup")
	log.InfoContext(context.Background(), "no fields")

	decoder := json.NewDecoder(&buf)
	for i := 0; i < 2; i++ {
		var line map[string]any
		require.NoError(t, decoder.Decode(&line))
		assert.NotContains(t, line, "requestID")
	}
}

func TestContextHandler_KeepsWithAttrsAndGroups(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(logger.NewContextHandler(slog.NewJSONHandler(&buf, nil)))
	log := logger.New(base.With("service", "api"))

	log.InfoContext(logger.WithFields(context.Background(), "requestID", "abc123"), "hello")

	line := decodeLine(t, &buf)
	assert.Equal(t, "api", line["service"])
	assert.Equal(t, "abc123", line["requestID"])
}
