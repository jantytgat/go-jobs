package orchestrator

type Queue interface {
	Push(t schedulerTick)
	Pop() (schedulerTick, error)
	Length() int
}
