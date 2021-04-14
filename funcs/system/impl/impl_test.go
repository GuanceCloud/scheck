package impl

import (
	"log"
	"os"
	"testing"
)

func TestUtmp(t *testing.T) {

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

func TestSockets(t *testing.T) {
	sockets, err := EnumProcessesOpenSockets(nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range sockets {
		log.Printf("pid=%d, fd=%d", s.PID, s.Fd)
	}
}
