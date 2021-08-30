package cgroup

import (
	"github.com/shirou/gopsutil/mem"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

const (
	KB     = 1024
	MB     = KB * 1024
	cpuMax = 100
	cpuMin = 0
	DefMEM = 200
)

var (
	l = logger.DefaultSLogger("cgroup")
)

func Run() {
	l = logger.SLogger("cgroup")

	if config.Cfg.Cgroup == nil || !config.Cfg.Cgroup.Enable {
		return
	}

	if !(float64(cpuMin) < config.Cfg.Cgroup.CPUMax && config.Cfg.Cgroup.CPUMax < float64(cpuMax)) ||
		!(float64(cpuMin) < config.Cfg.Cgroup.CPUMin && config.Cfg.Cgroup.CPUMin < float64(cpuMax)) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	if config.Cfg.Cgroup.CPUMax < config.Cfg.Cgroup.CPUMin {
		l.Errorf("CPUMin should less than CPUMax of the cgroup")
		return
	}
	if config.Cfg.Cgroup.MEM != -1 {
		if config.Cfg.Cgroup.MEM == 0 {
			config.Cfg.Cgroup.MEM = DefMEM
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
