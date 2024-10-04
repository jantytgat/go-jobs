package orchestrator

import (
	"time"

	"github.com/google/uuid"
)

type tick struct {
	uuid uuid.UUID
	time time.Time
}
