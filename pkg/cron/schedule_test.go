package cron

import (
	"testing"
	"time"
)

func TestNewSchedule(t *testing.T) {
	var tests = []struct {
		expression string
		success    bool
	}{
		{"*", false},
		{"* *", false},
		{"* * *", false},
		{"* * * *", false},
		{"a * * * *", false},

		{"* * * * *", true},
		{"* * * * * *", true},
		{"* * * * * * *", true},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			success := false

			_, err := NewSchedule(tt.expression)
			if err == nil {
				success = true
			}

			if success != tt.success {
				t.Errorf("unexpected result for %s with error: %s", tt.expression, err.Error())
			}

		})
	}
}

func TestSchedule_IsDueValid(t *testing.T) {
	var tests = []struct {
		expression string
		time       string
		wanted     bool
	}{
		{"* * * * * *", "20060102150405", true},
		{"5 * * * * *", "20060102150405", true},
		{"* 4 * * * *", "20060102150405", true},
		{"* * 15 * * *", "20060102150405", true},
		{"* * * 2 * *", "20060102150405", true},
		{"* * * * 1 *", "20060102150405", true},
		{"* * * * * 1 *", "20060102150405", true},
		// Different time value --> second == 0
		{"* * * * 1", "20060102150400", true},
		{"0 * * * * * *", "20060102150400", true},

		{"6 * * * * *", "20060102150405", false},
		{"* 5 * * * *", "20060102150405", false},
		{"* * 14 * * *", "20060102150405", false},
		{"* * * 3 * *", "20060102150405", false},
		{"* * * * 2 *", "20060102150405", false},
		{"* * * * * 2 *", "20060102150405", false},
		// Different time value --> second == 0
		{"* * * * 2", "20060102150400", false},
		{"5 * * * * * *", "20060102150400", false},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			i, _ := time.Parse("20060102150405", tt.time)
			s, _ := NewSchedule(tt.expression)
			if tt.wanted != s.IsDue(i) {
				t.Errorf("expected IsDue to be %t for %s", tt.wanted, tt.expression)
			}
		})
	}
}

func TestSchedule_ReplaceTemplates(t *testing.T) {
	var tests = []struct {
		expression string
		wanted     string
	}{
		{"@yearly", "0 0 1 1 *"},
		{"@annually", "0 0 1 1 *"},
		{"@monthly", "0 0 1 * *"},
		{"@weekly", "0 0 * * 0"},
		{"@daily", "0 0 * * *"},
		{"@hourly", "0 * * * *"},
		{"@everyminute", "* * * * *"},
		{"@5minutes", "*/5 * * * *"},
		{"@10minutes", "*/10 * * * *"},
		{"@15minutes", "*/15 * * * *"},
		{"@30minutes", "0,30 * * * *"},
		{"@everysecond", "* * * * * *"},

		{"a b c d e", "a b c d e"},
		{"1 2 3 4 5", "1 2 3 4 5"},
		{"random", "random"},
		{"@yearly @yearly", "@yearly @yearly"},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			s := Schedule{
				expression: tt.expression,
			}
			s.replaceTemplates()

			if s.expression != tt.wanted {
				t.Errorf("got %s, expected %s", s.expression, tt.wanted)
			}
		})
	}
}

func TestSchedule_Normalize(t *testing.T) {
	var tests = []struct {
		expression string
		wanted     string
	}{
		{"a  b c  d e", "A B C D E"},
		{"1  3  3  5  7  8", "1 3 3 5 7 8"},
		{"SUN 0 0 0 0", "0 0 0 0 0"},
		{"SUN  1 2 3 4", "0 1 2 3 4"},
		{"MON SUN 0 TUE sat", "1 0 0 2 6"},
		{"MON JAN 15 MAR cat", "1 1 15 3 CAT"},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			s := Schedule{
				expression: tt.expression,
			}
			s.normalize()

			if s.expression != tt.wanted {
				t.Errorf("got %s, expected %s", s.expression, tt.wanted)
			}
		})
	}
}

func TestSchedule_StandardizeValid(t *testing.T) {
	var tests = []struct {
		expression string
		wanted     bool
	}{
		{"* * * * *", true},
		{"* * * * * *", true},
		{"* * * * * 2006", true},
		{"* * * * * *", true},

		{"* * * *", false},
		{"* * * * * * * *", false},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			success := false

			s := Schedule{
				expression: tt.expression,
			}
			_, err := s.standardize()
			if err == nil {
				success = true
			}

			if success != tt.wanted {
				t.Errorf("unexpected result for %s, expected %t", tt.expression, tt.wanted)
			}
		})
	}
}

func TestSchedule_ParseValid(t *testing.T) {
	var tests = []struct {
		expression string
		wanted     bool
	}{
		{"* * * * *", true},
		{"* * * * * * *", true},

		{"* * * *", false},
		{"* * * * * * * *", false},
		{"a * * * *", false},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			success := false
			s := Schedule{
				expression: tt.expression,
			}
			if err := s.parse(); err == nil {
				success = true
			}

			if success != tt.wanted {
				t.Errorf("unexpected result for %s, expected %t", tt.expression, tt.wanted)
			}
		})
	}
}

func TestSchedule_String(t *testing.T) {
	var tests = []struct {
		expression string
		wanted     string
	}{
		{"* * * * *", "* * * * *"},
		{"* * * * * * *", "* * * * * * *"},
		{"@everysecond", "* * * * * *"},
	}

	for _, tt := range tests {
		t.Run(tt.expression, func(t *testing.T) {
			s, err := NewSchedule(tt.expression)
			if err != nil {
				t.Errorf("invalid expression %s", tt.expression)
			}

			if s.String() != tt.wanted {
				t.Errorf("unexpected output for %s, expected %s", tt.expression, tt.wanted)
			}
		})
	}
}
