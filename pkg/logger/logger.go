package logger

import (
	"context"
	"log/slog"
)

type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Debug(msg string, args ...any)
	InfoContext(ctx context.Context, msg string, args ...any)
	WarnContext(ctx context.Context, msg string, args ...any)
	ErrorContext(ctx context.Context, msg string, args ...any)
	DebugContext(ctx context.Context, msg string, args ...any)
}

type slogLogger struct {
	log *slog.Logger
}

func New(log *slog.Logger) Logger {
	return &slogLogger{log: log}
}

func (l *slogLogger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

func (l *slogLogger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}

func (l *slogLogger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

func (l *slogLogger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

func (l *slogLogger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.log.InfoContext(ctx, msg, args...)
}

func (l *slogLogger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.log.WarnContext(ctx, msg, args...)
}

func (l *slogLogger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.log.ErrorContext(ctx, msg, args...)
}

func (l *slogLogger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.log.DebugContext(ctx, msg, args...)
}
