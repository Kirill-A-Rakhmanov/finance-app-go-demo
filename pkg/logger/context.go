package logger

import (
	"context"
	"log/slog"
)

type fieldsKey struct{}

func WithFields(ctx context.Context, args ...any) context.Context {
	var record slog.Record
	record.Add(args...)

	fields := append([]slog.Attr(nil), fieldsFrom(ctx)...)
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, attr)
		return true
	})

	return context.WithValue(ctx, fieldsKey{}, fields)
}

func fieldsFrom(ctx context.Context) []slog.Attr {
	fields, _ := ctx.Value(fieldsKey{}).([]slog.Attr)

	return fields
}

type contextHandler struct {
	slog.Handler
}

func NewContextHandler(handler slog.Handler) slog.Handler {
	return &contextHandler{Handler: handler}
}

func (h *contextHandler) Handle(ctx context.Context, record slog.Record) error {
	if fields := fieldsFrom(ctx); len(fields) > 0 {
		record.AddAttrs(fields...)
	}

	return h.Handler.Handle(ctx, record)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
	return &contextHandler{Handler: h.Handler.WithGroup(name)}
}
