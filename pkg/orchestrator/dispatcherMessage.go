package orchestrator

import (
	"time"

	"github.com/jantytgat/go-jobs/pkg/job"
	"github.com/jantytgat/go-jobs/pkg/task"
)

type dispatcherMessage struct {
	job               job.Job
	handlerRepository *task.HandlerRepository
	trigger           time.Time
}
