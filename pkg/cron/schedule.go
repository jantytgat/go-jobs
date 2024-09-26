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

var reSpace = regexp.MustCompile(validSpace)

var cronWeekdayLiterals = strings.NewReplacer(
	"SUN", "0",
	"MON", "1",
	"TUE", "2",
	"WED", "3",
	"THU", "4",
	"FRI", "5",
	"SAT", "6")

var cronMonthLiterals = strings.NewReplacer(
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

var cronTemplates = map[string]string{
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

func NewSchedule(expression string) (Schedule, error) {
	s := Schedule{
		expression: expression,
		elements:   make([]element, 6),
	}
	s.replaceTemplates()
	s.normalize()
	err := s.parse()
	if err != nil {
		return Schedule{}, err
	}
	return s, nil
}

type Schedule struct {
	expression string
	elements   []element
}

func (s *Schedule) replaceTemplates() {
	// Replace template expression with a valid cron expression
	if t, ok := cronTemplates[s.expression]; ok {
		s.expression = t
	}
}

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

func (s *Schedule) standardize() ([]string, error) {
	segments := strings.Split(s.expression, " ")
	count := len(segments)

	// Expect at least 5 elements: minute, hour, day, month, weekday
	// Maximum 7 elements: second, minute, hour ,day, month, weekday, year
	if count < 5 || count > 7 {
		return nil, fmt.Errorf("invalid segment count, got %d, expected 5-7 elements separated by space", count)
	}

	if (count == 5) || (count == 6 && reYear.MatchString(segments[5])) {
		segments = append([]string{"0"}, segments...)
	}

	return segments, nil
}

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

func (s *Schedule) IsDue(t time.Time) bool {
	var o bool
	for _, e := range s.elements {
		if o = e.trigger(t); !o {
			break
		}
	}
	return o
}

func (s *Schedule) String() string {
	return s.expression
}
