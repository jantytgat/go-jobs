package cron

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	validSpace = `\s+`
)

var (
	reSpace = regexp.MustCompile(validSpace)

	cronWeekdayLiterals = strings.NewReplacer(
		"SUN", "0",
		"MON", "1",
		"TUE", "2",
		"WED", "3",
		"THU", "4",
		"FRI", "5",
		"SAT", "6")

	cronMonthLiterals = strings.NewReplacer(
		"JAN", "1",
		"FEB", "2",
		"MAR", "3",
		"APR", "4",
		"MAY", "5",
		"JUN", "6",
		"JUL", "7",
		"AUG", "8",
		"SEP", "9",
		"OCT", "10",
		"NOV", "11",
		"DEC", "12")

	cronTemplates = map[string]string{
		"@yearly":      "0 0 1 1 *",
		"@annually":    "0 0 1 1 *",
		"@monthly":     "0 0 1 * *",
		"@weekly":      "0 0 * * 0",
		"@daily":       "0 0 * * *",
		"@hourly":      "0 * * * *",
		"@everyminute": "* * * * *",
		"@5minutes":    "*/5 * * * *",
		"@10minutes":   "*/10 * * * *",
		"@15minutes":   "*/15 * * * *",
		"@30minutes":   "0,30 * * * *",
		"@everysecond": "* * * * * *",
	}
)

// Yearly returns a schedule which will trigger on the January 1st at 00:00
func Yearly() Schedule {
	s, _ := NewSchedule("@yearly")
	return s
}

// Monthly returns a schedule which will trigger on the first day of the month at 00:00
func Monthly() Schedule {
	s, _ := NewSchedule("@monthly")
	return s
}

// Weekly returns a schedule which will trigger each week on Sunday at 00:00
func Weekly() Schedule {
	s, _ := NewSchedule("@weekly")
	return s
}

// Daily returns a schedule which will trigger each day at 00:00
func Daily() Schedule {
	s, _ := NewSchedule("@daily")
	return s
}

// Hourly returns a schedule which will trigger at the top of each hour
func Hourly() Schedule {
	s, _ := NewSchedule("@hourly")
	return s
}

// EveryMinute returns a schedule which will trigger at the top of every minute
func EveryMinute() Schedule {
	s, _ := NewSchedule("@everyminute")
	return s
}

// EverySecond returns a schedule which will trigger at the top of every second
func EverySecond() Schedule {
	s, _ := NewSchedule("@everysecond")
	return s
}

// NewSchedule returns a Schedule based on the input expression.
// Returns an error if the expression cannot be parsed into separate elements.
func NewSchedule(expression string) (Schedule, error) {
	s := Schedule{
		expression: expression,
		elements:   make([]element, 6),
	}

	s.replaceTemplates()              // first replace all templates to literal cron schedules
	s.normalize()                     // normalize the expression to valid cron characters
	if err := s.parse(); err != nil { // parse the different elements in the schedule
		return Schedule{}, err
	}
	return s, nil
}

// Schedule defines a cron schedule based on a cron expression.
// Schedule should always be created using NewSchedule for proper initialization.
type Schedule struct {
	expression string
	elements   []element
}

// IsDue checks if input t matches the cron schedule defined by the expression.
func (s *Schedule) IsDue(t time.Time) bool {
	var o bool
	for _, e := range s.elements {
		if o = e.trigger(t); !o {
			break
		}
	}
	return o
}

// String returns the schedule expression as a string.
func (s *Schedule) String() string {
	return s.expression
}

// normalize replaces all strings and literals into cron characters.
// It does not parse or validate the expression!
func (s *Schedule) normalize() {
	// Replace all spaces with a single space
	s.expression = reSpace.ReplaceAllString(s.expression, " ")

	// Transform string to uppercase before replacing characters to numbers
	s.expression = strings.ToUpper(s.expression)

	// Replace weekday literals to numbers
	s.expression = cronWeekdayLiterals.Replace(s.expression)
	// Replace month literals to numbers
	s.expression = cronMonthLiterals.Replace(s.expression)
}

// parse converts the schedule expression into cron elements.
// Returns an error if the schedule expression cannot be converted into cron elements.
func (s *Schedule) parse() error {
	var (
		elements []string
		err      error
	)

	elements, err = s.standardize()
	if err != nil {
		return err
	}

	s.elements = make([]element, len(elements))

	for i, expression := range elements {
		var e element
		e, err = newElement(expression, position(i))
		if err != nil {
			return err
		}
		s.elements[i] = e
	}
	return nil
}

// replaceTemplates replaces the templates with their corresponding expression.
func (s *Schedule) replaceTemplates() {
	if t, ok := cronTemplates[s.expression]; ok {
		s.expression = t
	}
}

// standardize takes the expression as input and converts it into an array of strings.
// Returns an error the expression contains an invalid count of elements.
func (s *Schedule) standardize() ([]string, error) {
	segments := strings.Split(s.expression, " ")
	count := len(segments)

	// Expect at least 5 elements: minute, hour, day, month, weekday
	// Maximum 7 elements: second, minute, hour ,day, month, weekday, year
	if count < 5 || count > 7 {
		return nil, fmt.Errorf("invalid element count, got %d, expected 5-7 elements separated by space", count)
	}

	// If there are only 5 elements, prepend the expression with a 0 for the seconds position.
	// If there are 6 elements and the last element matches the regular expression for a year, prepend the expression with 0 for the seconds position.
	if (count == 5) || (count == 6 && reYear.MatchString(segments[5])) {
		segments = append([]string{"0"}, segments...)
	}

	return segments, nil
}
