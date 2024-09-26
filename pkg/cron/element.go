package cron

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	validSecondOrMinute = `(?:^(?:[1-5]?\d){1}$)|(?:^(?:[1-5]?\d)(?:,(?:[1-5]?\d))+$)|(?:^(?:[1-5]?\d)-(?:[1-5]?\d)$)|(?:^(?:\*/[2-6]|\*/10|\*/12|\*/15|\*/20|\*/30)$)`
	validHour           = `(?:^\*$^(?:(?:\d)|(?:1[0-9]{1})|(?:2[0-3]{1}))$)|(?:^(?:(?:\d)|(?:1[0-9]{1})|(?:2[0-3]{1}))(?:,(?:(?:\d)|(?:1[0-9]{1})|(?:2[0-3]{1})))*$)|(?:^(?:(?:\d)|(?:1[0-9]{1})|(?:2[0-3]{1}))-(?:(?:\d)|(?:1[0-9]{1})|(?:2[0-3]{1}))$)|(?:^(?:\*/[2-4]|\*/6|\*/8|\*/12)$)`
	validDayOfWeek      = `(?:^(?:(?:[0-6]{1}))$)|(?:^(?:(?:[0-6]{1}))(?:,(?:(?:[0-6]{1})))*$)|(?:^(?:(?:[0-6]{1}))-(?:(?:[0-6]{1}))$)`
	validDayOfMonth     = `(?:^(?:(?:[1-2]?[1-9]{1})|(?:3[0-1]{1}))$)|(?:^(?:(?:[1-2]?[1-9]{1})|(?:3[0-1]{1}))(?:,(?:(?:(?:[1-2]?[1-9]{1})|(?:3[0-1]{1}))))*$)|(?:^(?:(?:[1-2]?[1-9]{1})|(?:3[0-1]{1}))-(?:(?:[1-2]?[1-9]{1})|(?:3[0-1]{1}))$)`
	validMonth          = `(?:^(?:(?:[1-9])|(?:1[0-2]{1}))$)|(?:^(?:(?:[1-9])|(?:1[0-2]{1}))(?:,(?:(?:[1-9])|(?:1[0-2]{1})))*$)|(?:^(?:(?:[1-9])|(?:1[0-2]{1}))-(?:(?:[1-9])|(?:1[0-2]{1}))$)`
	validYear           = `(?:^\d+$)|(?:^(?:\d+)(?:,(?:\d+)+)*$)|(?:^\d+-\d+$)|(?:^\*\/\d+$)`
)

var reSecondOrMinute = regexp.MustCompile(validSecondOrMinute)
var reHour = regexp.MustCompile(validHour)
var reMonth = regexp.MustCompile(validMonth)
var reDayOfWeek = regexp.MustCompile(validDayOfWeek)
var reDayOfMonth = regexp.MustCompile(validDayOfMonth)
var reYear = regexp.MustCompile(validYear)

func newElement(expression string, position position) (element, error) {
	e := element{
		expression: expression,
		p:          position,
		q:          qualificationNone,
	}

	return e, e.validate()
}

type element struct {
	expression string
	p          position
	q          qualification
}

func (e *element) isDue(t int, inputs []int) bool {
	if len(inputs) == 0 {
		return false
	}

	switch e.q {
	case qualificationSimple:
		if inputs[0] == t {
			return true
		}
		return false
	case qualificationMulti:
		for _, i := range inputs {
			if i == t {
				return true
			}
		}
		return false
	case qualificationRange:
		if t >= inputs[0] && t <= inputs[1] {
			return true
		}
		return false
	case qualificationStep:
		if t%inputs[0] == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

func (e *element) parseExpression() ([]int, error) {
	var (
		s      []string
		output []int
		err    error
	)

	// No parsing must be done
	if e.expression == "*" {
		return output, nil
	}

	// Parsing depends on the qualification of the element
	switch e.q {
	case qualificationNone:
		return nil, fmt.Errorf("%s", "element has not been validated")
	case qualificationSimple:
		output = make([]int, 1)
		if output[0], err = strconv.Atoi(e.expression); err != nil {
			return output, err
		}
	case qualificationMulti:
		s = strings.Split(e.expression, ",")
		output = make([]int, len(s))
		for k, v := range s {
			if output[k], err = strconv.Atoi(v); err != nil {
				return output, err
			}
		}
	case qualificationRange:
		s = strings.Split(e.expression, "-")
		output = make([]int, 2)
		for k, v := range s {
			if output[k], err = strconv.Atoi(v); err != nil {
				return output, err
			}
		}
	case qualificationStep:
		s = strings.Split(e.expression, "*/")
		output = make([]int, 1)
		// Error handling is not required, as qualification is already parsed and step is an int
		if output[0], err = strconv.Atoi(s[1]); err != nil {
			return output, err
		}
	}

	// If there are multiple values, check if they are ascending
	if len(output) > 1 {
		for i := 0; i < len(output)-1; i++ {
			if output[i] > output[i+1] {
				return output, fmt.Errorf("invalid order of values in %s", e.p.String())
			}
		}
	}
	return output, nil
}

func (e *element) qualify() {
	if !strings.ContainsAny(e.expression, strings.Join(qualificationChars, "")) {
		e.q = qualificationSimple
		return
	}

	for i, char := range qualificationChars {
		if strings.Contains(e.expression, char) {
			// qualification starts at 2 = qualificationMulti
			e.q = qualification(i + 2)
			return
		}
	}
}

func (e *element) trigger(t time.Time) bool {
	if e.expression == "*" {
		return true
	}

	intervals, err := e.parseExpression()
	if err != nil {
		return false
	}

	var input int
	switch e.p {
	case positionSecond:
		input = t.Second()
	case positionMinute:
		input = t.Minute()
	case positionHour:
		input = t.Hour()
	case positionDay:
		input = t.Day()
	case positionMonth:
		input = int(t.Month())
	case positionWeekday:
		input = int(t.Weekday())
	case positionYear:
		input = t.Year()
	}
	return e.isDue(input, intervals)
}

func (e *element) validate() error {
	var err error

	// First qualify the expression
	if e.q == qualificationNone {
		e.qualify()
	}

	// No additional validation needed
	if e.expression == "*" {
		return nil
	}

	// Validate the element expression based on its position
	if err = e.validateExpressionForPosition(); err != nil {
		return err
	}

	// No additional validation is necessary for simple expressions
	if e.q == qualificationSimple {
		return nil
	}

	// There can be no mixture of different qualifications as these are imposed by regex
	if _, err = e.parseExpression(); err != nil {
		return err
	}
	return nil
}

func (e *element) validateExpressionForPosition() error {
	// Validate element by position
	var s []string
	switch e.p {
	case positionSecond:
		s = reSecondOrMinute.FindStringSubmatch(e.expression)
	case positionMinute:
		s = reSecondOrMinute.FindStringSubmatch(e.expression)
	case positionHour:
		s = reHour.FindStringSubmatch(e.expression)
	case positionDay:
		s = reDayOfMonth.FindStringSubmatch(e.expression)
	case positionMonth:
		s = reMonth.FindStringSubmatch(e.expression)
	case positionWeekday:
		s = reDayOfWeek.FindStringSubmatch(e.expression)
	case positionYear:
		s = reYear.FindStringSubmatch(e.expression)
	}

	// No match found
	if len(s) == 0 {
		return fmt.Errorf("invalid expression in %s", e.p.String())
	}
	return nil
}
