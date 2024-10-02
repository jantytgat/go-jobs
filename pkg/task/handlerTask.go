package task

func NewHandlerTask(t Task, p *Pipeline) HandlerTask {
	return HandlerTask{
		Task:     t,
		Pipeline: p,
		ChResult: make(chan HandlerTaskResult),
	}
}

type HandlerTask struct {
	Task     Task
	Pipeline *Pipeline
	ChResult chan HandlerTaskResult
}

type HandlerTaskResult struct {
	Status Status
	Error  error
}
