package job

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func NewMemoryCatalog() *MemoryCatalog {
	return &MemoryCatalog{
		jobs: make(map[uuid.UUID]Job),
	}
}

type MemoryCatalog struct {
	jobs map[uuid.UUID]Job

	mux sync.RWMutex
}

func (c *MemoryCatalog) Add(job Job) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.jobs[job.Uuid]; ok {
		return fmt.Errorf("job with uuid %s already exists", job.Uuid)
	}

	c.jobs[job.Uuid] = job
	return nil
}

func (c *MemoryCatalog) All() map[uuid.UUID]Job {
	c.mux.RLock()
	defer c.mux.RUnlock()

	jobs := make(map[uuid.UUID]Job, len(c.jobs))
	for k, v := range c.jobs {
		jobs[k] = v
	}
	return jobs
}

func (c *MemoryCatalog) Count() int {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return len(c.jobs)
}

func (c *MemoryCatalog) Delete(uuid uuid.UUID) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.jobs[uuid]; !ok {
		return fmt.Errorf("job with uuid %s does not exist", uuid)
	}

	delete(c.jobs, uuid)
	return nil
}

func (c *MemoryCatalog) Get(uuid uuid.UUID) (Job, error) {
	c.mux.RLock()
	defer c.mux.RUnlock()

	if _, ok := c.jobs[uuid]; !ok {
		return Job{}, fmt.Errorf("job with uuid %s does not exist", uuid)
	}

	return c.jobs[uuid], nil
}

func (c *MemoryCatalog) Statistics() CatalogStatistics {
	var enabled, disabled int
	jobs := c.All()
	for _, v := range jobs {
		switch v.Enabled {
		case true:
			enabled++
		case false:
			disabled++
		}
	}

	return CatalogStatistics{
		Count:         enabled + disabled,
		EnabledCount:  enabled,
		DisabledCount: disabled,
	}
}

func (c *MemoryCatalog) Update(job Job) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.jobs[job.Uuid]; !ok {
		return fmt.Errorf("job with uuid %s does not exist", job.Uuid)
	}
	c.jobs[job.Uuid] = job
	return nil
}
