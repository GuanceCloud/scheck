package cgroup

import (
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/mem"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

var (
	l  = logger.DefaultSLogger("cgroup")
	MB = int64(1024 * 1024)
)

func Run() {
	l = logger.SLogger("cgroup")

	if config.Cfg.Cgroup == nil || !config.Cfg.Cgroup.Enable {
		return
	}

	if !(0 < config.Cfg.Cgroup.CPUMax && config.Cfg.Cgroup.CPUMax < 100) ||
		!(0 < config.Cfg.Cgroup.CPUMin && config.Cfg.Cgroup.CPUMin < 100) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	if config.Cfg.Cgroup.CPUMax < config.Cfg.Cgroup.CPUMin {
		l.Errorf("CPUMin should less than CPUMax of the cgroup")
		return
	}
	if config.Cfg.Cgroup.MEM != -1 {
		if config.Cfg.Cgroup.MEM == 0 {
			config.Cfg.Cgroup.MEM = 200
		}

		vm, err := mem.VirtualMemory()
		if err != nil {
			l.Warn("MEM VirtualMemory err=%v", err)
			return
		}

		available := vm.Available / uint64(MB)
		if uint64(config.Cfg.Cgroup.MEM) > available {
			l.Errorf("MEM should less than Available")
			return
		}
	}

	start()
}

var blocks = make([][1024 * 512]byte, 0)

func memTest() {
	time.Sleep(time.Second)
	for {
		var m1 runtime.MemStats
		runtime.ReadMemStats(&m1)

		if m1.Alloc > uint64(1024*1024*30) {
			os.Exit(0)
		} else {
			blocks = append(blocks, [1024 * 512]byte{})
		}
		time.Sleep(time.Second * 2)
	}

}
func bToMb(b uint64) uint64 {
	return b / uint64(MB)
}
