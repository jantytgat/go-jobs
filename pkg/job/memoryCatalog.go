package job

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

func NewMemoryCatalog() *MemoryCatalog {
	return &MemoryCatalog{
		jobs:    make(map[uuid.UUID]Job),
		results: make(map[uuid.UUID][]Result),
	}
}

type MemoryCatalog struct {
	jobs    map[uuid.UUID]Job
	results map[uuid.UUID][]Result

	mux sync.Mutex
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

func (c *MemoryCatalog) AddResult(result Result) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.results[result.Uuid]; !ok {
		c.results[result.Uuid] = make([]Result, 0)
	}

	c.results[result.Uuid] = append(c.results[result.Uuid], result)
}

func (c *MemoryCatalog) All() map[uuid.UUID]Job {
	c.mux.Lock()
	defer c.mux.Unlock()

	jobs := make(map[uuid.UUID]Job)
	for k, v := range c.jobs {
		jobs[k] = v
	}
	return jobs
}

func (c *MemoryCatalog) AllResults() map[uuid.UUID][]Result {
	c.mux.Lock()
	defer c.mux.Unlock()

	results := make(map[uuid.UUID][]Result)
	for k, v := range c.results {
		results[k] = v
	}
	return results
}

func (c *MemoryCatalog) Count() int {
	c.mux.Lock()
	defer c.mux.Unlock()
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
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.jobs[uuid]; !ok {
		return Job{}, fmt.Errorf("job with uuid %s does not exist", uuid)
	}

	return c.jobs[uuid], nil
}

func (c *MemoryCatalog) GetResults(uuid uuid.UUID) ([]Result, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if _, ok := c.results[uuid]; !ok {
		return nil, fmt.Errorf("results for job with uuid %s do not exist", uuid)
	}

	return c.results[uuid], nil
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

	var resultCount int
	results := c.AllResults()
	for _, v := range results {
		resultCount += len(v)
	}

	return CatalogStatistics{
		Count:         enabled + disabled,
		EnabledCount:  enabled,
		DisabledCount: disabled,
		ResultCount:   resultCount,
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
