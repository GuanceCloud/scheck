package checker

import (
	"testing"
)

func TestParse(t *testing.T) {
	cronStr := `* */1 * * *`

	it := checkRunTime(cronStr)
	t.Log(it)
}

func TestShowFuncs(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"file"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ShowFuncs()
		})
	}
}
