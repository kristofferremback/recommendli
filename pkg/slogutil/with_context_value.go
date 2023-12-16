package slogutil

import (
	"context"
	"log/slog"
)

type ctxkey string

const attrsKey ctxkey = "slogutil/attrs"

func WithAttrs(ctx context.Context, newAttrs ...slog.Attr) context.Context {
	return context.WithValue(ctx, attrsKey, append(GetAttrs(ctx), newAttrs...))
}

func GetAttrs(ctx context.Context) []slog.Attr {
	attrs, _ := ctx.Value(attrsKey).([]slog.Attr)
	return attrs
}

var _ slog.Handler = (*CtxHandler)(nil)

type CtxHandler struct {
	slog.Handler
}

func (h *CtxHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(GetAttrs(ctx)...)
	return h.Handler.Handle(ctx, r)
}

func (h *CtxHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.Handler.Enabled(ctx, level)
}

func (h *CtxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CtxHandler{h.Handler.WithAttrs(attrs)}
}

func (h *CtxHandler) WithGroup(name string) slog.Handler {
	return &CtxHandler{h.Handler.WithGroup(name)}
}
