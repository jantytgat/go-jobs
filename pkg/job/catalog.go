package job

import "github.com/google/uuid"

type Catalog interface {
	Add(job Job) error
	All() map[uuid.UUID]Job
	Count() int
	Delete(uuid uuid.UUID) error
	Get(uuid uuid.UUID) (Job, error)
	Statistics() CatalogStatistics
	Update(job Job) error
}
