package task

import (
	"context"
	"time"
)

type Task interface {
	Name() string
	DefaultHandler() Handler
	DefaultHandlerPool(ctx context.Context) *HandlerPool
	Handler(timeout time.Duration) Handler
	HandlerPool(ctx context.Context, timeout time.Duration) *HandlerPool
}
