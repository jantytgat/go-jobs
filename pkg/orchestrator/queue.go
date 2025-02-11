package orchestrator

type Queue interface {
	Push(t SchedulerTick)
	Pop() (SchedulerTick, error)
	Length() int
}
