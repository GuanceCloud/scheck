package cgroup

import (
	"os"
	"runtime"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

	"github.com/containerd/cgroups"
	specs "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/shirou/gopsutil/cpu"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
)

const (
	LevelLow  = "low"
	LevelHigh = "high"
)

var (
	period    = uint64(global.CgroupPeriod) // 1 second
	hundred   = 100
	hundredf  = float64(hundred)
	hundred1f = 100.0
)

func start() {
	high := config.Cfg.Cgroup.CPUMax * float64(runtime.NumCPU()) / hundredf
	low := config.Cfg.Cgroup.CPUMin * float64(runtime.NumCPU()) / hundredf

	quotaHigh := int64(float64(period) * high)
	quotaLow := int64(float64(period) * low)
	memLimit := int64(config.Cfg.Cgroup.MEM) * global.MB
	swap := memLimit
	pid := os.Getpid()

	l.Infof("with %d CPU, set CPU limimt %.2f%%", runtime.NumCPU(), float64(quotaLow)/float64(period)*hundred1f)

	control, err := cgroups.New(cgroups.V1, cgroups.StaticPath("/scheck"),
		&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &quotaLow,
			},
			Memory: &specs.LinuxMemory{
				Limit: &memLimit,
				Swap:  &swap,
			},
		})
	if err != nil {
		l.Errorf("failed of new cgroup: %s", err)
		return
	}
	defer control.Delete() // nolint:errcheck

	if err := control.Add(cgroups.Process{Pid: pid}); err != nil {
		l.Errorf("faild of add cgroup: %s", err)
		return
	}

	l.Infof("add PID %d to cgroup", pid)

	level := LevelLow
	waitNum := 0
	maxWaitNum := 3
	cupMaxHigh := 95
	timeSleep := 3

	for {
		percpu, err := getCPUPercent(time.Second * time.Duration(timeSleep))
		if err != nil {
			l.Debug(err)
			continue
		}

		var q int64
		// 当前 cpu 使用率 + 设定的最大使用率 超过 95% 时，将使用 low 模式
		// 否则如果连续 3 次判断小于 95%，则使用 high 模式
		if float64(cupMaxHigh) < percpu+high {
			if level == LevelLow {
				continue
			}
			q = quotaLow
			level = LevelLow
		} else {
			if level == LevelHigh {
				continue
			}
			if waitNum < maxWaitNum {
				waitNum++
				continue
			}
			q = quotaHigh
			level = LevelHigh
			waitNum = 0
		}

		err = control.Update(&specs.LinuxResources{
			CPU: &specs.LinuxCPU{
				Period: &period,
				Quota:  &q,
			}})
		if err != nil {
			l.Debugf("failed of update cgroup: %s", err)
			continue
		}
		l.Debugf("switch to quota %.2f%%", float64(q)/float64(period)*hundred1f)
	}
}

func getCPUPercent(interval time.Duration) (float64, error) {
	percent, err := cpu.Percent(interval, false)
	if err != nil {
		return 0, err
	}
	if len(percent) == 0 {
		return 0, nil
	}
	return percent[0] / hundredf, nil
}
