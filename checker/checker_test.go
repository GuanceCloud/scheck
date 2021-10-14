package checker

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestParse(t *testing.T) {
	cronStr := `* */1 * * *`

	it := checkRunTime(cronStr)
	log.Println(it)
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
