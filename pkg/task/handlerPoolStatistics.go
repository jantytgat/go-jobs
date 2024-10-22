package task

type HandlerPoolStatistics struct {
	ActiveWorkers  int
	IdleWorkers    int
	Workers        int
	MaxWorkers     int
	TasksProcessed int
}
