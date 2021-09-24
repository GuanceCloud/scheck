package impl

import (
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
		t.Fatal(err)
	}
	defer f.Close()

	items, err := ParseUtmp(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range items {
		t.Log("username=%s, tty=%v, host=%v", item.User, item.Device, item.Host)
	}
}

func TestSockets(t *testing.T) {
	if runtime.GOOS == global.OSWindows {
		return
	}
	sockets, err := EnumProOpenSockets(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, s := range sockets {
		t.Logf("pid=%d, fd=%d", s.PID, s.Fd)
	}
}
