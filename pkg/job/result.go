package job

import (
	"time"

	"github.com/google/uuid"

	"github.com/jantytgat/go-jobs/pkg/task"
)

type Result struct {
	Uuid        uuid.UUID
	RunUuid     uuid.UUID
	Trigger     time.Time
	RunTime     time.Duration
	TaskResults []task.Result
	Error       error
}
