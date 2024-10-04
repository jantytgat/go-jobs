package cron

import "testing"

func TestExpressionType_String(t *testing.T) {
	var (
		result []string
		wanted = qualificationStrings
	)

	for i := 0; i < len(wanted); i++ {
		result = append(result, qualification(i).String())
	}

	for j := 0; j < len(wanted); j++ {
		if result[j] != wanted[j] {
			t.Errorf("invalid string: got %s expected %s", result[j], wanted[j])
		}
	}
}
