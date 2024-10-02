package task

type Status int

func (s Status) String() string {
	return [...]string{
		"none",
		"pending",
		"success",
		"canceled",
		"error",
	}[s]

}

const (
	StatusNone Status = iota
	StatusPending
	StatusSuccess
	StatusCanceled
	StatusError
)
