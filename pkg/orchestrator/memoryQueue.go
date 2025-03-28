package orchestrator

import (
	"errors"
	"sync"
)

func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{}
}

type MemoryQueue struct {
	queue []SchedulerTick
	mux   sync.Mutex
}

func (q *MemoryQueue) Push(t SchedulerTick) {
	q.mux.Lock()
	defer q.mux.Unlock()

	q.queue = append(q.queue, t)
}

func (q *MemoryQueue) Pop() (SchedulerTick, error) {
	if q.Length() == 0 {
		return SchedulerTick{}, errors.New("empty queue")
	}
	q.mux.Lock()
	defer q.mux.Unlock()

	tick := q.queue[0]
	q.queue = q.queue[1:]
	return tick, nil
}

func (q *MemoryQueue) Length() int {
	q.mux.Lock()
	defer q.mux.Unlock()
	return len(q.queue)
}
