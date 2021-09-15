package impl

import (
	"log"
	"os"
	"runtime"
	"testing"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

func TestUtmp(t *testing.T) {
	if runtime.GOOS == global.OSWindows {
		return
	}
	f, err := os.Open(`/var/run/utmp`)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	items, err := ParseUtmp(f)
	if err != nil {
		log.Println(err)
		return
	}
	for _, item := range items {
		log.Printf("username=%s, tty=%v, host=%v", item.User, item.Device, item.Host)
	}
}

func TestSockets(t *testing.T) {
	if runtime.GOOS == global.OSWindows {
		return
	}
	sockets, err := EnumProOpenSockets(nil)
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range sockets {
		log.Printf("pid=%d, fd=%d", s.PID, s.Fd)
	}
}
