package orchestrator

import (
	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/cron"
)

type schedulerUpdate struct {
	uuid     uuid.UUID
	enabled  bool
	schedule cron.Schedule
}
