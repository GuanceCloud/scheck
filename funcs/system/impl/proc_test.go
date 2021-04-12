package impl

import (
	"log"
	"testing"
)

func TestProc(t *testing.T) {
	pss, err := GetProcesses()
	if err != nil {
		log.Fatal(err)
	}
	for _, p := range pss {
		log.Printf("%s", p.Name)
	}
}
