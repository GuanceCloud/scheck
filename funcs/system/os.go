// Package system export fot lua
package system

import (
	"os"
	"time"

	diskutil "github.com/shirou/gopsutil/disk"
	hostutil "github.com/shirou/gopsutil/host"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	apiLock = make(chan int, 1)
	api     = map[string]lua.LGFunction{
		"hostname":  Hostname,
		"time_zone": Zone,
		"uptime":    Uptime,
		"mounts":    Mounts,
		"uname":     Uname,
		"sc_sleep":  Sleep,
		"sc_ticker": Ticker,
		"sc_log":    ScLog,
	}
)

func Loader(l *lua.LState) int {
	apiLock <- 0
	for name, fn := range loadOtherOS() {
		api[name] = fn
	}

	t := l.NewTable()
	mod := l.SetFuncs(t, api)
	<-apiLock
	l.Push(mod)
	return 1
}

func Hostname(l *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(name))
	return 1
}

func Zone(l *lua.LState) int {
	z, _ := time.Now().Zone()
	l.Push(lua.LString(z))
	return 1
}

func Uname(l *lua.LState) int {
	info, err := hostutil.Info()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	var tbl lua.LTable
	tbl.RawSetString("platform", lua.LString(info.Platform))
	tbl.RawSetString("family", lua.LString(info.PlatformFamily))
	tbl.RawSetString("platform_version", lua.LString(info.PlatformVersion))
	tbl.RawSetString("os", lua.LString(info.OS))
	tbl.RawSetString("arch", lua.LString(info.KernelArch))
	tbl.RawSetString("kernel_version", lua.LString(info.KernelVersion))

	l.Push(&tbl)
	return 1
}

func Uptime(l *lua.LState) int {
	info, err := hostutil.Info()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LNumber(info.Uptime))
	return 1
}

func Mounts(l *lua.LState) int {
	parts, err := diskutil.Partitions(true)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	var ms lua.LTable
	for _, p := range parts {
		var pt lua.LTable
		pt.RawSetString("device", lua.LString(p.Device))
		pt.RawSetString("path", lua.LString(p.Mountpoint))
		pt.RawSetString("type", lua.LString(p.Fstype))
		pt.RawSetString("flags", lua.LString(p.Opts))

		ms.Append(&pt)
	}

	l.Push(&ms)
	return 1
}

func Sleep(l *lua.LState) int {
	num := 0
	lv := l.Get(1)
	if lv.Type() != lua.LTNil {
		if lv.Type() != lua.LTNumber {
			l.TypeError(1, lua.LTNumber)
			return lua.MultRet
		}
		num = int(lv.(lua.LNumber))
	}
	time.Sleep(time.Duration(num) * time.Second)
	return 0
}

func Ticker(l *lua.LState) int {
	chanN := 1
	intN := 2
	var interval time.Duration
	scChan := l.ToChannel(chanN)
	lv := l.Get(intN)
	if lv.Type() != lua.LTNil {
		interval = 1 * time.Second
	} else {
		if lv.Type() == lua.LTNumber {
			interval = time.Duration(int(lv.(lua.LNumber))) * time.Second
		} else {
			interval = 1 * time.Second
		}
	}
	go func(interval time.Duration) {
		timer1 := time.NewTicker(interval)
		for v := range timer1.C {
			scChan <- lua.LString(v.String())
		}
	}(interval)
	return 0
}

func ScLog(l *lua.LState) int {
	str := ""
	lv := l.Get(1)
	if lv.Type() != lua.LTNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		}
		str = string(lv.(lua.LString))
	}
	loger := logger.DefaultSLogger("lua")
	loger.Info(str)
	return 0
}
