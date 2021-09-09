package cgroup

import (
	"github.com/shirou/gopsutil/mem"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

const (
	cpuMax = 100
	cpuMin = 0
	DefMEM = 200
)

var (
	l = logger.DefaultSLogger("cgroup")
)

func Run(cgroup *config.Cgroup) {
	l = logger.SLogger("cgroup")

	if cgroup == nil || !cgroup.Enable {
		return
	}

	if !(float64(cpuMin) < cgroup.CPUMax && cgroup.CPUMax < float64(cpuMax)) ||
		!(float64(cpuMin) < cgroup.CPUMin && cgroup.CPUMin < float64(cpuMax)) {
		l.Errorf("CPUMax and CPUMin should be in range of (0.0, 100.0)")
		return
	}

	if cgroup.CPUMax < cgroup.CPUMin {
		l.Errorf("CPUMin should less than CPUMax of the cgroup")
		return
	}
	if cgroup.MEM != -1 {
		if cgroup.MEM == 0 {
			cgroup.MEM = DefMEM
		}

		vm, err := mem.VirtualMemory()
		if err != nil {
			l.Warn("MEM VirtualMemory err=%v", err)
			return
		}

		available := vm.Available / uint64(global.MB)
		if uint64(cgroup.MEM) > available {
			l.Errorf("MEM should less than Available")
			return
		}
	}

	start()
}
