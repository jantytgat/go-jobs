package task

import (
	"context"
	"log/slog"
	"time"
)

type Task interface {
	Name() string
	DefaultHandler() Handler
	DefaultHandlerPool(ctx context.Context) *HandlerPool
	Handler(timeout time.Duration) Handler
	HandlerPool(ctx context.Context, timeout time.Duration) *HandlerPool
}

func LogTaskAttr(t Task) slog.Attr {
	return slog.Group("task",
		slog.String("name", t.Name()))
}
