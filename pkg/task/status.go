package task

const (
	StatusNone Status = iota
	StatusPending
	StatusSuccess
	StatusCanceled
	StatusError
)

var StatusStrings = []string{"none", "pending", "success", "canceled", "error"}

type Status int

func (s Status) String() string {
	return StatusStrings[s]

}
