package job

import (
	"log/slog"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

func NewJob(uuid uuid.UUID, name string, schedule cron.Schedule) Job {
	return Job{
		Uuid:             uuid,
		Name:             name,
		Schedule:         schedule,
		Enabled:          false,
		LimitConcurrency: true,
		MaxConcurrency:   1,
		LimitRuns:        false,
		MaxRuns:          0,
	}
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
}

func (j Job) Disable() Job {
	j.Enabled = false
	return j
}

func (j Job) Enable() Job {
	j.Enabled = true
	return j
}

func (j Job) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("Uuid", j.Uuid.String()),
		slog.String("Name", j.Name))
}

func (j Job) LogAttrs() []slog.Attr {
	return []slog.Attr{slog.Any("job", j)}
}

func (j Job) WithMaxConcurrency(i int) Job {
	j.LimitConcurrency = true
	j.MaxConcurrency = i
	return j
}

func (j Job) WithMaxRuns(i int) Job {
	j.LimitRuns = true
	j.MaxRuns = i
	return j
}

func (j Job) WithNoConcurrency() Job {
	j.LimitConcurrency = true
	j.MaxConcurrency = 1
	return j
}

func (j Job) WithUnlimitedConcurrency() Job {
	j.LimitConcurrency = false
	return j
}

func (j Job) WithUnlimitedRuns() Job {
	j.LimitRuns = false
	return j
}
