package orchestrator

import "github.com/jantytgat/go-jobs/pkg/task"

type Statistics struct {
	HandlerPoolStatistics map[string]task.HandlerPoolStatistics
	QueueLength           int
	JobsSuccess           int
	JobsFailed            int
}
