package job

import "github.com/google/uuid"

type Catalog interface {
	Add(job Job) error
	AddResult(result Result)
	All() map[uuid.UUID]Job
	AllResults() map[uuid.UUID][]Result
	Count() int
	CountResults(uuid uuid.UUID) int
	Delete(uuid uuid.UUID) error
	Get(uuid uuid.UUID) (Job, error)
	GetResults(uuid uuid.UUID) ([]Result, error)
	Statistics() CatalogStatistics
	Update(job Job) error
}
