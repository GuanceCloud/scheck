package impl

import (
	"log"
	"os"
	"testing"
)

func TestProc(t *testing.T) {

	f, err := os.Open(`/var/run/utmp`)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	items, err := ParseUtmp(f)
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range items {
		log.Printf("username=%s, tty=%v, host=%v", item.User, item.Device, item.Host)
	}

}
