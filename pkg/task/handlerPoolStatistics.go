package task

type HandlerPoolStatistics struct {
	ActiveWorkers                float64
	IdleWorkers                  float64
	Workers                      float64
	RecycledWorkers              float64
	MaxWorkers                   float64
	TasksIngested                float64
	TasksProcessedStatusNone     float64
	TasksProcessedStatusPending  float64
	TasksProcessedStatusSuccess  float64
	TasksProcessedStatusCanceled float64
	TasksProcessedStatusError    float64
	TasksWaiting                 float64
}
