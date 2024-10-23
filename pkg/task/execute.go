package task

import (
	"context"
	"log/slog"
)

func Execute(ctx context.Context, l *slog.Logger, task Task, r *HandlerRepository) (Result, error) {
	pipeline := NewPipeline(l)
	chResults := make(chan HandlerResult)
	if err := r.Execute(ctx, NewHandlerTaskWithChannel(task, pipeline, chResults)); err != nil {
		return Result{}, err
	}

	// Wait for the result of the current task
	for {
		select {
		case <-ctx.Done():
			return Result{}, ctx.Err()
		case result := <-chResults:
			return Result{
				Status: result.Status,
				Error:  result.Error,
			}, nil
		}
	}
}

func ExecuteSequence(ctx context.Context, l *slog.Logger, tasks []Task, r *HandlerRepository) ([]Result, error) {
	pipeline := NewPipeline(l)
	chResults := make(chan HandlerResult)
	results := make([]Result, len(tasks))
	for i, task := range tasks {
		// If the HandlerPool cannot be found in the HandlerRepository, the repository will first try to register
		// the pool based on the task. If the registration fails, the HandlerRepository will return an error.
		// We cannot proceed with the execution of the sequence, so we return the error to be handled by the caller.
		// The handler pool will send the result of the task to s.chResults.
		// Any data that needs to be passed on through the sequence of tasks is stored in the pipeline by the task handler.
		if err := r.Execute(ctx, NewHandlerTaskWithChannel(task, pipeline, chResults)); err != nil {
			return results, err
		}

		// Wait for the result of the current task
		for {
			select {
			case <-ctx.Done():
				return results, ctx.Err()
			case result := <-chResults:
				results[i] = Result{
					Status: result.Status,
					Error:  result.Error,
				}
			}
		}
	}
	return results, nil
}
