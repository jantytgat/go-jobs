package task

import "testing"

func TestStatus_String(t *testing.T) {
	var (
		result []string
		wanted = StatusStrings
	)

	for i := 0; i < len(wanted); i++ {
		result = append(result, Status(i).String())
	}

	for j := 0; j < len(wanted); j++ {
		if result[j] != wanted[j] {
			t.Errorf("invalid string: got %s expected %s", result[j], wanted[j])
		}
	}
}
