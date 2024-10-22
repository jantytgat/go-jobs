package job

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
	"github.com/jantytgat/go-jobs/pkg/task"
)

func New(uuid uuid.UUID, name string, schedule cron.Schedule, tasks []task.Task, opts ...Option) Job {
	j := Job{
		Uuid:             uuid,
		Name:             name,
		Schedule:         schedule,
		Enabled:          true,
		LimitConcurrency: true,
		MaxConcurrency:   1,
		LimitRuns:        false,
		MaxRuns:          0,
		Tasks:            tasks,
	}

	for _, opt := range opts {
		opt(&j)
	}

	return j
}

type Job struct {
	Uuid             uuid.UUID
	Name             string
	Schedule         cron.Schedule
	Enabled          bool
	LimitConcurrency bool
	MaxConcurrency   int
	LimitRuns        bool
	MaxRuns          int
	Tasks            []task.Task
}

func (j *Job) Disable() {
	j.Enabled = false
}

func (j *Job) Enable() {
	j.Enabled = true
}

func (j *Job) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Uuid", j.Uuid.String()),
		slog.String("Name", j.Name))
}

func (j *Job) LogAttrs() []slog.Attr {
	return []slog.Attr{slog.Any("job", j)}
}
