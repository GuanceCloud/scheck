package luaext

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	lua "github.com/yuin/gopher-lua"

	diskutil "github.com/shirou/gopsutil/disk"
	hostutil "github.com/shirou/gopsutil/host"
	processutil "github.com/shirou/gopsutil/process"
)

func zone(l *lua.LState) int {
	z, _ := time.Now().Zone()
	l.Push(lua.LString(z))
	return 1
}

func uptime(l *lua.LState) int {
	info, err := hostutil.Info()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LNumber(info.Uptime))
	return 1
}

func kernelInfo(l *lua.LState) int {
	var kernel lua.LTable

	if data, err := ioutil.ReadFile(`/proc/cmdline`); err == nil {
		args := strings.Split(string(data), " ")
		additionalArguments := ""
		for _, arg := range args {
			if len(arg) > 11 && arg[0:11] == "BOOT_IMAGE=" {
				kernel.RawSetString("path", lua.LString(arg[11:]))
			} else if len(arg) > 5 && arg[0:5] == "root=" {
				kernel.RawSetString("device", lua.LString(arg[5:]))
			} else {
				if additionalArguments != "" {
					additionalArguments += " "
				}
				additionalArguments += arg
			}
		}
		if additionalArguments != "" {
			kernel.RawSetString("arguments", lua.LString(strings.TrimSpace(additionalArguments)))
		}
	}

	if data, err := ioutil.ReadFile(`/proc/version`); err == nil {
		details := strings.Split(string(data), " ")
		if len(details) > 2 && details[1] == "version" {
			kernel.RawSetString("version", lua.LString(details[2]))
		}
	}

	l.Push(&kernel)
	return 1
}

func kernelModules(l *lua.LState) int {

	var result lua.LTable

	if data, err := ioutil.ReadFile(`/proc/modules`); err == nil {
		mods := strings.Split(string(data), "\n")
		for _, mod := range mods {
			mod = strings.TrimSpace(mod)
			parts := strings.FieldsFunc(mod, func(r rune) bool {
				if r == ' ' {
					return true
				}
				return false
			})
			if len(parts) < 5 {
				continue
			}
			for idx, p := range parts {
				if len(p) > 0 && p[len(p)-1] == ',' {
					parts[idx] = p[0 : len(p)-1]
				}
			}
			var mt lua.LTable
			mt.RawSetString("name", lua.LString(parts[0]))
			mt.RawSetString("size", lua.LString(parts[1]))
			mt.RawSetString("used_by", lua.LString(parts[2]))
			mt.RawSetString("status", lua.LString(parts[3]))
			mt.RawSetString("address", lua.LString(parts[4]))
			result.Append(&mt)
		}
	}

	l.Push(&result)
	return 1
}

func users(l *lua.LState) int {

	us, err := getUserDetail("")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable

	for _, u := range us {
		var ut lua.LTable
		ut.RawSetString("username", lua.LString(u.name))
		ut.RawSetString("uid", lua.LString(u.uid))
		ut.RawSetString("gid", lua.LString(u.gid))
		ut.RawSetString("home", lua.LString(u.home))
		ut.RawSetString("shell", lua.LString(u.shell))
		result.Append(&ut)
	}

	l.Push(&result)
	return 1
}

func loggedInUsers(l *lua.LState) int {
	return 1
}

func shadow(l *lua.LState) int {

	type shadowInfo struct {
		status     string
		lastChange int64
		expireAt   int64
	}

	file, err := os.Open("/etc/shadow")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	defer file.Close()

	var result lua.LTable
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				l.RaiseError("%s", err)
				return lua.MultRet
			}
		}

		parts := strings.Split(line, ":")
		if len(parts) < 9 {
			continue
		}

		var status string
		lastChange := -1
		expire := -1

		psw := parts[1]
		if psw == "" {
			status = "empty"
		} else {
			if psw == "!!" {
				status = "not-set"
			} else if psw[:1] == "*" || psw[:1] == "!" || psw[:1] == "x" {
				status = "locked"
			} else {
				status = "active"
			}
		}

		if d, err := strconv.Atoi(parts[2]); err == nil && d > 0 {
			lastChange = d
		}
		if d, err := strconv.Atoi(parts[8]); err == nil && d > 0 {
			expire = d
		}

		var ut lua.LTable
		ut.RawSetString("username", lua.LString(parts[0]))
		ut.RawSetString("password_status", lua.LString(status))
		ut.RawSetString("last_change", lua.LNumber(lastChange))
		ut.RawSetString("expire", lua.LNumber(expire))
		result.Append(&ut)
	}

	l.Push(&result)
	return 1
}

func ulimitInfo(l *lua.LState) int {

	var result lua.LTable

	limitsResourceMap := []struct {
		name string
		id   int
	}{
		{"cpu", syscall.RLIMIT_CPU},
		{"fsize", syscall.RLIMIT_FSIZE},
		{"data", syscall.RLIMIT_DATA},
		{"stack", syscall.RLIMIT_STACK},
		{"core", syscall.RLIMIT_CORE},
		{"nofile", syscall.RLIMIT_NOFILE},
		{"as", syscall.RLIMIT_AS},
	}

	for _, r := range limitsResourceMap {
		var rLimit syscall.Rlimit
		err := syscall.Getrlimit(r.id, &rLimit)
		if err == nil {
			var item lua.LTable
			item.RawSetString("type", lua.LString(r.name))
			item.RawSetString("soft_limit", lua.LNumber(rLimit.Cur))
			item.RawSetString("hard_limit", lua.LNumber(rLimit.Max))
			result.Append(&item)
		} else {
			log.Printf("[error] failt ot getrlimit, %s", err)
		}
	}
	l.Push(&result)
	return 1
}

func mounts(l *lua.LState) int {
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

func processes(l *lua.LState) int {

	pslist, err := processutil.Processes()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	var result lua.LTable
	for _, p := range pslist {
		name, _ := p.Name()
		cmdline, _ := p.Cmdline()
		cpu, _ := p.CPUPercent()
		exe, _ := p.Exe()
		uids, _ := p.Uids()
		gids, _ := p.Gids()
		times, _ := p.Times()
		starttime, _ := p.CreateTime()
		threads, _ := p.Threads()
		status, _ := p.Status()
		var pt lua.LTable
		pt.RawSetString("pid", lua.LNumber(p.Pid))
		pt.RawSetString("name", lua.LString(name))
		pt.RawSetString("cmdline", lua.LString(cmdline))
		pt.RawSetString("percent_processor_time", lua.LNumber(cpu))
		pt.RawSetString("path", lua.LString(exe))
		if len(uids) > 0 {
			pt.RawSetString("uid", lua.LNumber(uids[0]))
		}
		if len(gids) > 0 {
			pt.RawSetString("gid", lua.LNumber(gids[0]))
		}
		if times != nil {
			pt.RawSetString("system_time", lua.LNumber(times.System))
			pt.RawSetString("user_time", lua.LNumber(times.User))
			pt.RawSetString("nice", lua.LNumber(times.Nice))
		}
		pt.RawSetString("start_time", lua.LNumber(starttime/1000))
		pt.RawSetString("threads", lua.LNumber(len(threads)))
		pt.RawSetString("state", lua.LString(status))
		result.Append(&pt)
	}

	l.Push(&result)
	return 1
}

func shellHistory(l *lua.LState) int {

	targetUser := ""
	lv := l.Get(1)
	if lv.Type() != lua.LTNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		} else {
			targetUser = lv.String()
		}
	}

	users, err := getUserDetail(targetUser)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	} else {
		if len(users) == 0 {
			l.RaiseError("user '%s' not exists", targetUser)
			return lua.MultRet
		}
	}

	shellHistoryFiles := []string{
		".bash_history",
		".zsh_history",
		".zhistory",
		".history",
		".sh_history",
	}

	var result lua.LTable
	for _, u := range users {
		if u.home == "" {
			continue
		}
		if u.shell == "/bin/false" || strings.HasSuffix(u.shell, "/nologin") {
			continue
		}

		for _, f := range shellHistoryFiles {

			cmds, err := genShellHistoryFromFile(filepath.Join(u.home, f))
			if err != nil {
				l.RaiseError("%s", err)
				return lua.MultRet
			}

			for _, cmd := range cmds {
				var item lua.LTable
				item.RawSetString("uid", lua.LString(u.uid))
				item.RawSetString("history_file", lua.LString(filepath.Join(u.home, f)))
				item.RawSetString("command", lua.LString(cmd.command))
				item.RawSetString("time", lua.LNumber(cmd.time))
				result.Append(&item)
			}
		}
	}

	l.Push(&result)
	return 1
}

func last(l *lua.LState) int {

	// targetUser := ""
	// limit := -1
	// lv := l.Get(1)
	// if lv.Type() != lua.LTNil {
	// 	if lv.Type() != lua.LTString {
	// 		l.TypeError(1, lua.LTString)
	// 		return lua.MultRet
	// 	} else {
	// 		targetUser = lv.String()
	// 	}
	// }

	// lv = l.Get(2)
	// if lv.Type() != lua.LTNil {
	// 	if lv.Type() != lua.LTNumber {
	// 		l.TypeError(2, lua.LTNumber)
	// 		return lua.MultRet
	// 	} else {
	// 		limit = int(lv.(lua.LNumber))
	// 	}
	// }

	buf := bytes.NewBufferString("")
	cmd := exec.Command("last")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	reader := bufio.NewReader(buf)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		fields := strings.FieldsFunc(line, func(r rune) bool {
			if r == ' ' {
				return true
			}
			return false
		})

		if len(fields) < 5 {
			continue
		}

		log.Printf("%v", len(fields))

	}

	return 1
}
