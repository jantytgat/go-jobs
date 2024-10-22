package task

import "sync"

type IntercomStatistics struct {
	ErrorCount int
}

type Intercom struct {
	mux    sync.Mutex
	errors []error
}

func (i *Intercom) Errors() []error {
	i.mux.Lock()
	defer i.mux.Unlock()
	return i.errors
}

func (i *Intercom) Statistics() IntercomStatistics {
	i.mux.Lock()
	defer i.mux.Unlock()

	return IntercomStatistics{
		ErrorCount: len(i.errors),
	}
}
