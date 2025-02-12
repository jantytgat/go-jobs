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
	handlerCtx, handlerCancel := context.WithTimeout(ctx, h.timeout)
	defer handlerCancel()

	chHandlerOutput := make(chan error, 1)

	go func(ctx context.Context, t Task, p *Pipeline, chOut chan error) {
		defer close(chOut)
		chOut <- h.execute(ctx, t, p)
	}(handlerCtx, t, p, chHandlerOutput)

	select {
	case <-handlerCtx.Done():
		return StatusCanceled, fmt.Errorf("handler context canceled for %s: %w", h.Name, handlerCtx.Err())
	case err := <-chHandlerOutput:
		if err != nil {
			return StatusError, err
		}
		return StatusSuccess, nil
	}
}
