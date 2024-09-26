package cron

import (
	"strconv"
	"testing"
	"time"
)

func TestNewElement(t *testing.T) {
	for i := positionSecond; i < positionYear+1; i++ {
		_, err := newElement("*", i)
		if err != nil {
			t.Errorf("invalid element, got %s", err.Error())
		}
	}
}

func TestNewElement_SecondMinuteValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"0"},
		{"11"},
		{"59"},

		{"0,1"},
		{"11,48"},

		{"0-1"},
		{"11-59"},

		{"*/2"},
		{"*/30"},
	}

	for _, tt := range tests {
		for i := range []position{positionSecond, positionMinute} {
			t.Run(position(i).String()+"_"+tt.expression, func(t *testing.T) {
				e, err := newElement(tt.expression, position(i))
				if err != nil {
					t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
				}
			})
		}
	}
}

func TestNewElement_SecondMinuteInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"a"},

		{"60"},
		{"110"},

		{"22,60"},
		{"24,64"},
		{"31,21"},
		{"1,a"},

		{"28-60"},
		{"35-65"},
		{"36-1"},

		{"*/0"},
		{"*/1"},
		{"*/31"},

		{"#"},
		{"*/a"},
	}

	for _, tt := range tests {
		for i := range []position{positionSecond, positionMinute} {
			t.Run(position(i).String()+"_"+tt.expression, func(t *testing.T) {
				e, err := newElement(tt.expression, position(i))
				if err == nil {
					t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
				}
			})

		}
	}
}

func TestNewElement_HourValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"0"},
		{"11"},
		{"23"},

		{"0,1"},
		{"9,11"},
		{"11,23"},

		{"0-1"},
		{"5-8"},
		{"15-22"},

		{"*/2"},
		{"*/3"},
		{"*/4"},
		{"*/6"},
		{"*/8"},
		{"*/12"},
	}

	for _, tt := range tests {
		t.Run(positionHour.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionHour)
			if err != nil {
				t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
			}
		})
	}
}

func TestNewElement_HourInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"a"},

		{"24"},
		{"110"},

		{"11,64"},
		{"12,10"},
		{"1,a"},

		{"1-24"},
		{"13-65"},
		{"6-1"},

		{"*/0"},
		{"*/1"},
		{"*/9"},
		{"*/13"},
	}

	for _, tt := range tests {
		t.Run(positionHour.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionHour)
			if err == nil {
				t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
			}
		})
	}
}

func TestNewElement_DayValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"5"},
		{"11"},
		{"23"},
		{"31"},

		{"1,4"},
		{"9,11"},
		{"11,23"},
		{"21,31"},

		{"1-6"},
		{"5-8"},
		{"15-31"},
	}

	for _, tt := range tests {
		t.Run(positionDay.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionDay)
			if err != nil {
				t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
			}
		})
	}
}

func TestNewElement_DayInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"0"},
		{"a"},

		{"32"},
		{"105"},

		{"11,32"},
		{"0,10"},
		{"1,a"},

		{"0-24"},
		{"13-65"},
		{"6-1"},

		{"*/0"},
		{"*/1"},
		{"*/9"},
		{"*/13"},
	}

	for _, tt := range tests {
		t.Run(positionDay.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionDay)
			if err == nil {
				t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
			}
		})
	}
}

func TestNewElement_MonthValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"5"},
		{"12"},

		{"1,4"},
		{"9,11"},
		{"11,12"},

		{"1-6"},
		{"5-8"},
		{"6-12"},
	}

	for _, tt := range tests {
		t.Run(positionMonth.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionMonth)
			if err != nil {
				t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
			}
		})
	}
}

func TestNewElement_MonthInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"0"},
		{"a"},

		{"13"},
		{"105"},

		{"11,32"},
		{"0,10"},
		{"1,a"},

		{"0-24"},
		{"13-65"},
		{"6-1"},

		{"*/0"},
		{"*/1"},
		{"*/9"},
		{"*/13"},
	}

	for _, tt := range tests {
		t.Run(positionMonth.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionMonth)
			if err == nil {
				t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
			}
		})
	}
}

func TestNewElement_WeekdayValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"0"},
		{"5"},
		{"6"},

		{"1,4"},
		{"0,6"},

		{"1-3"},
		{"0-6"},
	}

	for _, tt := range tests {
		t.Run(positionWeekday.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionWeekday)
			if err != nil {
				t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
			}
		})
	}
}

func TestNewElement_WeekdayInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"a"},

		{"7"},
		{"105"},

		{"0,10"},
		{"1,a"},

		{"0-24"},
		{"6-1"},

		{"*/0"},
		{"*/1"},
		{"*/9"},
		{"*/13"},
	}

	for _, tt := range tests {
		t.Run(positionWeekday.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionWeekday)
			if err == nil {
				t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
			}
		})
	}
}

func TestNewElement_YearValid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{"*"},
		{"2000"},

		{"1000,4000"},

		{"1000-3000"},

		{"*/1"},
		{"*/22"},
		{"*/333"},
		{"*/4444"},
	}

	for _, tt := range tests {
		t.Run(positionYear.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionYear)
			if err != nil {
				t.Errorf("invalid element, got %s with error %s", e.q.String(), err.Error())
			}
		})
	}
}

func TestNewElement_YearInvalid(t *testing.T) {
	var tests = []struct {
		expression string
	}{
		{""},
		{"a"},

		{"1,a"},
		{"22,a"},
		{"333,a"},
		{"a,1"},
		{"a,22"},
		{"a,333"},
		{"22,11"},

		{"1-a"},
		{"22-a"},
		{"333-a"},
		{"a-1"},
		{"a-22"},
		{"a-333"},
		{"33-11"},

		{"*/a"},
		{"*/"},
	}

	for _, tt := range tests {
		t.Run(positionYear.String()+"_"+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, positionYear)
			if err == nil {
				t.Errorf("invalid %s, got invalid expression %s", e.p.String(), e.expression)
			}
		})
	}
}

func TestElement_parseExpressionValid(t *testing.T) {
	var tests = []struct {
		expression string
		q          qualification
	}{
		{"*", qualificationSimple},
		{"0", qualificationSimple},

		{"0,20,31,100", qualificationMulti},
		{"0,20,31,100", qualificationMulti},
		{"0,20,31,100", qualificationMulti},
		{"0,20,31,100", qualificationMulti},

		{"0-200", qualificationRange},
		{"0-200", qualificationRange},
	}

	for _, tt := range tests {
		t.Run(tt.q.String()+"_"+tt.expression, func(t *testing.T) {
			e := element{
				expression: tt.expression,
				q:          tt.q,
			}
			if _, err := e.parseExpression(); err != nil {
				t.Errorf("unexpected error for expression %s with qualification %s", tt.expression, tt.q.String())
			}
		})
	}
}

func TestElement_parseExpressionInvalid(t *testing.T) {
	var tests = []struct {
		expression string
		q          qualification
	}{
		{"a", qualificationSimple},

		{"0,a,31,100", qualificationMulti},
		{"0,20,31,*", qualificationMulti},

		{"0-a", qualificationRange},
		{"a-200", qualificationRange},
	}

	for _, tt := range tests {
		t.Run(tt.q.String()+"_"+tt.expression, func(t *testing.T) {
			e := element{
				expression: tt.expression,
				q:          tt.q,
			}
			if _, err := e.parseExpression(); err == nil {
				t.Errorf("expected error for expression %s with qualification %s", tt.expression, tt.q.String())
			}
		})
	}
}

func TestElement_isDueValid(t *testing.T) {
	var tests = []struct {
		q      qualification
		inputs []int
		t      int
	}{
		{qualificationSimple, []int{0}, 0},
		{qualificationSimple, []int{200}, 200},

		{qualificationMulti, []int{0, 20, 31, 100}, 0},
		{qualificationMulti, []int{0, 20, 31, 100}, 20},
		{qualificationMulti, []int{0, 20, 31, 100}, 31},
		{qualificationMulti, []int{0, 20, 31, 100}, 100},

		{qualificationRange, []int{0, 200}, 0},
		{qualificationRange, []int{0, 200}, 31},

		{qualificationStep, []int{2}, 60},
	}

	for _, tt := range tests {
		t.Run(tt.q.String()+"_"+strconv.Itoa(tt.t), func(t *testing.T) {
			e := element{
				q: tt.q,
			}
			if !e.isDue(tt.inputs, tt.t) {
				t.Errorf("expected true for %d", tt.t)
			}
		})
	}
}

func TestElement_isDueInvalid(t *testing.T) {
	var tests = []struct {
		q      qualification
		inputs []int
		t      int
	}{
		{qualificationNone, nil, 0},
		{qualificationNone, []int{0}, 0},

		{qualificationSimple, nil, 0},
		{qualificationSimple, []int{}, 200},
		{qualificationSimple, []int{0}, 200},

		{qualificationMulti, nil, 0},
		{qualificationMulti, []int{0, 10}, 200},
		{qualificationMulti, []int{0, 10, 300}, 200},

		{qualificationRange, nil, 0},
		{qualificationRange, []int{0, 10}, 200},
		{qualificationRange, []int{0, 10, 300}, 200},

		{qualificationStep, []int{7}, 60},
	}

	for _, tt := range tests {
		t.Run(tt.q.String()+"_"+strconv.Itoa(tt.t), func(t *testing.T) {
			e := element{
				q: tt.q,
			}
			if e.isDue(tt.inputs, tt.t) {
				t.Errorf("expected false for %d", tt.t)
			}
		})
	}
}

func TestElement_TriggerValid(t *testing.T) {
	var tests = []struct {
		p          position
		expression string
	}{
		{positionSecond, "*"},
		{positionSecond, "5"},

		{positionMinute, "*"},
		{positionMinute, "4"},

		{positionHour, "*"},
		{positionHour, "15"},

		{positionDay, "*"},
		{positionDay, "2"},

		{positionMonth, "*"},
		{positionMonth, "1"},

		{positionWeekday, "*"},
		{positionWeekday, "1"},

		{positionYear, "*"},
		{positionYear, "2006"},
	}

	// INPUT TIME: Monday, January 2, 2006, at 15:04:05
	i, _ := time.Parse("20060102150405", "20060102150405")

	for _, tt := range tests {
		t.Run(tt.p.String()+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, tt.p)
			if err != nil {
				t.Errorf("invalid element for %s with expression %s", tt.p.String(), tt.expression)
			}

			if !e.trigger(i) {
				t.Errorf("expected true for value %s with expression %s", i.String(), tt.expression)
			}
		})
	}
}

func TestElement_TriggerInvalid(t *testing.T) {
	var tests = []struct {
		p          position
		expression string
	}{
		{positionSecond, "6"},

		{positionMinute, "9"},

		{positionHour, "18"},

		{positionDay, "5"},

		{positionMonth, "9"},

		{positionWeekday, "2"},

		{positionYear, "1"},
		{positionYear, "22"},
		{positionYear, "333"},
	}

	// INPUT TIME: Monday, January 2, 2006, at 15:04:05
	i, _ := time.Parse("20060102150405", "20060102150405")

	for _, tt := range tests {
		t.Run(tt.p.String()+tt.expression, func(t *testing.T) {
			e, err := newElement(tt.expression, tt.p)
			if err != nil {
				t.Errorf("invalid element for %s with expression %s", tt.p.String(), tt.expression)
			}

			if e.trigger(i) {
				t.Errorf("expected false for value %s with expression %s", i.String(), tt.expression)
			}
		})
	}
}

func TestElement_TriggerInvalid2(t *testing.T) {
	var tests = []struct {
		p          position
		expression string
	}{
		{positionSecond, "6"},
	}

	// INPUT TIME: Monday, January 2, 2006, at 15:04:05
	i, _ := time.Parse("20060102150405", "20060102150405")

	for _, tt := range tests {
		t.Run(tt.p.String()+tt.expression, func(t *testing.T) {
			e := element{
				expression: tt.expression,
				p:          tt.p,
				q:          qualificationNone,
			}

			if e.trigger(i) {
				t.Errorf("expected false for value %s with expression %s", i.String(), tt.expression)
			}
		})
	}
}
