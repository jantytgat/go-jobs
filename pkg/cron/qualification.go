package cron

var qualificationChars = []string{",", "-", "/"}

type qualification int

func (q qualification) String() string {
	return [...]string{"invalid", "simple", "multi", "range", "step"}[q]
}

const (
	qualificationNone qualification = iota
	qualificationSimple
	qualificationMulti
	qualificationRange
	qualificationStep
)
