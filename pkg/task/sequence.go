package task

import (
	"context"
	"log/slog"
	"sync"
)

func NewSequence(l *slog.Logger, tasks []Task) Sequence {
	return Sequence{
		tasks:     tasks,
		pipeline:  NewPipeline(l),
		chResults: make(chan HandlerResult),
	}
}

type Sequence struct {
	tasks   []Task
	results []Result

	pipeline  *Pipeline
	chResults chan HandlerResult

	mux sync.Mutex
}

func (s *Sequence) Execute(ctx context.Context, r *HandlerRepository) error {
	// Copy the tasks in a local variable to avoid data races
	tasks := make([]Task, len(s.tasks))
	s.mux.Lock()
	copy(tasks, s.tasks)
	s.mux.Unlock()

	results := make([]Result, len(tasks))
	for i, task := range tasks {
		// If the HandlerPool cannot be found in the HandlerRepository, the repository will first try to register
		// the pool based on the task. If the registration fails, the HandlerRepository will return an error.
		// We cannot proceed with the execution of the sequence, so we return the error to be handled by the caller.
		// The handler pool will send the result of the task to s.chResults.
		// Any data that needs to be passed on through the sequence of tasks is stored in the pipeline by the task handler.
		if err := r.Execute(ctx, NewHandlerTaskWithChannel(task, s.pipeline, s.chResults)); err != nil {
			return err
		}

		// Wait for the result of the current task
		result := <-s.chResults
		results[i] = Result{
			Status: result.Status,
			Error:  result.Error,
		}
	}
	s.mux.Lock()
	s.results = results
	s.mux.Unlock()
	return nil
}
