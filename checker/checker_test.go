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
