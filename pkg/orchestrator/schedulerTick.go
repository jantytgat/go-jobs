package orchestrator

import (
	"time"

	"github.com/google/uuid"
)

type schedulerTick struct {
	uuid uuid.UUID
	time time.Time
}
