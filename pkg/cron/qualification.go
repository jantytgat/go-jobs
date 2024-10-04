package cron

const (
	qualificationNone   qualification = iota
	qualificationSimple               // qualification does not contain any of the qualification characters
	qualificationMulti                // qualification contains a comma
	qualificationRange                // qualification contains a dash
	qualificationStep                 // qualification contains a slash
)

var (
	qualificationStrings = []string{"none", "simple", "multi", "range", "step"}
	qualificationChars   = []string{",", "-", "/"}
)

type qualification int

func (q qualification) String() string {
	return qualificationStrings[q]
}
