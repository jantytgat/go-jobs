package task

func NewHandlerTask(t Task, p *Pipeline) HandlerTask {
	return HandlerTask{
		Task:     t,
		Pipeline: p,
		ChResult: make(chan HandlerResult),
	}
}

func NewHandlerTaskWithChannel(t Task, p *Pipeline, ch chan HandlerResult) HandlerTask {
	return HandlerTask{
		Task:     t,
		Pipeline: p,
		ChResult: ch,
	}
}

type HandlerTask struct {
	Task     Task
	Pipeline *Pipeline
	ChResult chan HandlerResult
}
