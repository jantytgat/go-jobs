package task

type HandlerPoolStatistics struct {
	ActiveWorkers  float64
	IdleWorkers    float64
	Workers        float64
	MaxWorkers     float64
	TasksIngested  float64
	TasksProcessed float64
	TasksWaiting   float64
}
