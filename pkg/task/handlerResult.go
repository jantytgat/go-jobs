package task

type HandlerResult struct {
	Task   Task
	Status Status
	Error  error
}
