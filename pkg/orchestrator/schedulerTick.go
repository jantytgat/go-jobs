package orchestrator

import (
	"time"

	"github.com/google/uuid"
)

type SchedulerTick struct {
	uuid uuid.UUID
	time time.Time
}
