package orchestrator

import "context"

func NewRunner(ctx context.Context, maxRunners int) (*runner, chan tick) {
	chRunner := make(chan tick, maxRunners)
	return &runner{
		ctx:     ctx,
		chQueue: chRunner,
	}, chRunner
}

type runner struct {
	ctx     context.Context
	chQueue chan tick
}
