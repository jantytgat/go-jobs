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

// newElement returns a new cron element based on the input expression and position.
// Returns an error if the expression cannot be validated for the position.
func newElement(expression string, position position) (element, error) {
	e := element{
		expression: expression,
		p:          position,
		q:          qualificationNone,
	}

	return e, e.validate() // element has pointer receivers, validate() will update the qualification on return
}

// Element specifies the expression for a specific position in the cron schedule definition.
type element struct {
	expression string
	p          position
	q          qualification
}

// isDue checks if the value of t matches with any of the inputs, based on the qualification of the element.
func (e *element) isDue(t int, inputs []int) bool {
	// Fail-fast check to make sure there is input to check against.
	if len(inputs) == 0 {
		return false
	}

	switch e.q {
	case qualificationSimple: // only a single value is possible for the input
		if inputs[0] == t {
			return true
		}
		return false
	case qualificationMulti: // multiple values are possible, we can match on any input value
		for _, i := range inputs {
			if i == t {
				return true
			}
		}
		return false
	case qualificationRange: // input must be in a range between two input values
		if t >= inputs[0] && t <= inputs[1] {
			return true
		}
		return false
	case qualificationStep: // value must return 0 when performing a mod division by the input
		if t%inputs[0] == 0 {
			return true
		}
		return false
	default:
		return false
	}
}

// parseExpression parses the expression into an array of int which can be used to check if the element is due.
// An error will be returned if the value of the expression cannot be parsed into an array of int.
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
	case qualificationNone: // validate the expression in case the element hasn't got a proper qualification
		if err = e.validate(); err != nil {
			return nil, err
		}
		return e.parseExpression()
	case qualificationSimple: // parsing should return a single value
		output = make([]int, 1)
		if output[0], err = strconv.Atoi(e.expression); err != nil {
			return output, err
		}
	case qualificationMulti: // expression has multiple values separated by a comma
		s = strings.Split(e.expression, ",")
		output = make([]int, len(s))
		for k, v := range s {
			if output[k], err = strconv.Atoi(v); err != nil {
				return output, err
			}
		}
	case qualificationRange: // expression defines a ranges of values between two numbers
		s = strings.Split(e.expression, "-")
		output = make([]int, 2)
		for k, v := range s {
			if output[k], err = strconv.Atoi(v); err != nil {
				return output, err
			}
		}
	case qualificationStep:
		output = make([]int, 1)
		s = strings.Split(e.expression, "*/")
		if e.expression == s[0] {
			return output, fmt.Errorf("expression %s does not start with */", e.expression)
		}

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

// qualify takes the expression for the element and detects the qualification (simple, multi, range, step)
func (e *element) qualify() {
	if !strings.ContainsAny(e.expression, strings.Join(qualificationChars, "")) {
		e.q = qualificationSimple
		return
	}

	for i, char := range qualificationChars {
		if strings.Contains(e.expression, char) {
			// offset qualification for complex expressions, starts at i=2: qualificationMulti
			e.q = qualification(i + 2)
			return
		}
	}
}

// trigger checks if input t aligns with the expression for the element, depending on the position and qualification of the element.
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

// validate ensures that the element has a valid expression set for its position.
// Returns an error if the expression cannot be validated for its position.
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

// validateExpressionForPosition validates the expression of the element against the predefined regular expression for that position.
// Returns an error if no match was found for the regular expression.
func (e *element) validateExpressionForPosition() error {
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
