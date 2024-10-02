package task

import (
	"context"
	"fmt"
	"time"
)

func NewHandler(name string, timeout time.Duration, f func(ctx context.Context, t Task, p *Pipeline) error) Handler {
	return Handler{
		Name:    name,
		timeout: timeout,
		execute: f,
	}
}

type Handler struct {
	Name    string
	timeout time.Duration
	execute func(ctx context.Context, t Task, p *Pipeline) error
}

func (h Handler) Execute(ctx context.Context, t Task, p *Pipeline) (Status, error) {
	chHandlerOutput := make(chan error, 1)
	handlerCtx, handlerCancel := context.WithTimeout(ctx, h.timeout)
	defer handlerCancel()

	go func(ctx context.Context, t Task, p *Pipeline) {
		chHandlerOutput <- h.execute(ctx, t, p)
	}(handlerCtx, t, p)

	select {
	case <-ctx.Done():
		return StatusCanceled, fmt.Errorf("handler context canceled: %w", ctx.Err())
	case <-handlerCtx.Done():
		return StatusCanceled, fmt.Errorf("handler timeout for %s after %d seconds", h.Name, h.timeout)
	case err := <-chHandlerOutput:
		if err != nil {
			return StatusError, err
		}
		return StatusSuccess, nil
	}
}
