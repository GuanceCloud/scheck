package system

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system/impl"

	diskutil "github.com/shirou/gopsutil/disk"
	hostutil "github.com/shirou/gopsutil/host"
)

func (p *provider) hostname(l *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(name))
	return 1
}

func (p *provider) zone(l *lua.LState) int {
	z, _ := time.Now().Zone()
	l.Push(lua.LString(z))
	return 1
}

func (p *provider) uptime(l *lua.LState) int {
	info, err := hostutil.Info()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LNumber(info.Uptime))
	return 1
}

func (p *provider) kernelInfo(l *lua.LState) int {
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

func (p *provider) kernelModules(l *lua.LState) int {

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

func (p *provider) users(l *lua.LState) int {

	us, err := impl.GetUserDetail("")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable

	for _, u := range us {
		var ut lua.LTable
		ut.RawSetString("username", lua.LString(u.Name))
		uid := -1
		gid := -1
		if n, err := strconv.Atoi(u.UID); err == nil {
			uid = n
		}
		if n, err := strconv.Atoi(u.GID); err == nil {
			uid = n
		}
		ut.RawSetString("uid", lua.LNumber(uid))
		ut.RawSetString("gid", lua.LNumber(gid))
		ut.RawSetString("directory", lua.LString(u.Home))
		ut.RawSetString("shell", lua.LString(u.Shell))
		result.Append(&ut)
	}

	l.Push(&result)
	return 1
}

func loggedInUsers(l *lua.LState) int {
	return 1
}

func (p *provider) shadow(l *lua.LState) int {

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

func (p *provider) ulimitInfo(l *lua.LState) int {

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
			v := ""
			if int(rLimit.Cur) == syscall.RLIM_INFINITY {
				v = "unlimited"
			} else {
				v = fmt.Sprintf("%v", rLimit.Cur)
			}
			item.RawSetString("soft_limit", lua.LString(v))

			if int(rLimit.Max) == syscall.RLIM_INFINITY {
				v = "unlimited"
			} else {
				v = fmt.Sprintf("%v", rLimit.Max)
			}
			item.RawSetString("hard_limit", lua.LString(v))
			result.Append(&item)
		} else {
			log.Errorf("fail ot getrlimit, %s", err)
		}
	}
	l.Push(&result)
	return 1
}

func (p *provider) mounts(l *lua.LState) int {
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

func (p *provider) processes(l *lua.LState) int {

	pslist, err := impl.GetProcesses()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	var result lua.LTable
	for _, p := range pslist {
		var pt lua.LTable
		pt.RawSetString("pid", lua.LNumber(p.Pid))
		pt.RawSetString("parent", lua.LNumber(p.Parent))
		pt.RawSetString("path", lua.LString(p.Path))
		pt.RawSetString("name", lua.LString(p.Name))
		pt.RawSetString("pgroup", lua.LString(p.PGroup))
		pt.RawSetString("state", lua.LString(p.State))
		pt.RawSetString("nice", lua.LString(p.Nice))
		pt.RawSetString("threads", lua.LNumber(p.Threads))
		pt.RawSetString("cmdline", lua.LString(p.Cmdline))
		pt.RawSetString("cwd", lua.LString(p.Cwd))
		pt.RawSetString("root", lua.LString(p.Root))
		pt.RawSetString("uid", lua.LString(p.UID))
		pt.RawSetString("euid", lua.LString(p.EUID))
		pt.RawSetString("suid", lua.LString(p.SUID))
		pt.RawSetString("gid", lua.LString(p.GID))
		pt.RawSetString("egid", lua.LString(p.EGID))
		pt.RawSetString("sgid", lua.LString(p.SGID))
		pt.RawSetString("on_disk", lua.LNumber(p.OnDisk))
		pt.RawSetString("resident_size", lua.LNumber(p.ResidentSize))
		pt.RawSetString("total_size", lua.LNumber(p.TotalSize))
		pt.RawSetString("user_time", lua.LNumber(p.UserTime))
		pt.RawSetString("system_time", lua.LNumber(p.SystemTime))
		pt.RawSetString("start_time", lua.LNumber(p.StartTime))
		pt.RawSetString("disk_bytes_read", lua.LNumber(p.DiskBytesRead))
		pt.RawSetString("disk_bytes_written", lua.LNumber(p.DiskBytesWritten))

		result.Append(&pt)
	}

	l.Push(&result)
	return 1
}

func (p *provider) processOpendFiles(l *lua.LState) int {

	var pids []int
	lv := l.Get(1)
	if lv != lua.LNil {
		if lv.Type() != lua.LTNumber {
			l.TypeError(1, lua.LTNumber)
			return lua.MultRet
		}
		pids = append(pids, int(lv.(lua.LNumber)))
	}

	fds, err := impl.EnumProcessesFds(pids)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable

	for _, fd := range fds {
		var t lua.LTable
		t.RawSetString("pid", lua.LNumber(fd.PID))
		t.RawSetString("fd", lua.LString(fd.Fd))
		t.RawSetString("path", lua.LString(fd.LinkPath))
		result.Append(&t)
	}

	l.Push(&result)
	return 1
}

func (p *provider) shellHistory(l *lua.LState) int {

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

	users, err := impl.GetUserDetail(targetUser)
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
		if u.Home == "" {
			continue
		}
		if u.Shell == "/bin/false" || strings.HasSuffix(u.Shell, "/nologin") {
			continue
		}

		for _, f := range shellHistoryFiles {

			cmds, err := impl.GenShellHistoryFromFile(filepath.Join(u.Home, f))
			if err != nil {
				l.RaiseError("%s", err)
				return lua.MultRet
			}

			for _, cmd := range cmds {
				var item lua.LTable
				uid := -1
				if n, err := strconv.Atoi(u.UID); err == nil {
					uid = n
				}
				item.RawSetString("uid", lua.LNumber(uid))
				item.RawSetString("history_file", lua.LString(filepath.Join(u.Home, f)))
				item.RawSetString("command", lua.LString(cmd.Command))
				item.RawSetString("time", lua.LNumber(cmd.Time))
				result.Append(&item)
			}
		}
	}

	l.Push(&result)
	return 1
}

func (p *provider) parseUtmpFile(file string) (*lua.LTable, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var result lua.LTable
	utmps, err := impl.ParseUtmp(f)
	if err != nil {
		return nil, err
	}
	for _, utmp := range utmps {
		var item lua.LTable
		item.RawSetString("pid", lua.LNumber(utmp.Pid))
		item.RawSetString("username", lua.LString(utmp.User))
		item.RawSetString("tty", lua.LString(utmp.Device))
		item.RawSetString("host", lua.LString(utmp.Host))
		item.RawSetString("type", lua.LNumber(utmp.Type))
		item.RawSetString("time", lua.LNumber(utmp.Time))
		result.Append(&item)
	}
	return &result, nil
}

func (p *provider) last(l *lua.LState) int {

	utmps, err := p.parseUtmpFile("/var/log/wtmp")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(utmps)
	return 1
}

func (p *provider) lastb(l *lua.LState) int {

	utmps, err := p.parseUtmpFile("/var/log/btmp")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(utmps)
	return 1
}

func (p *provider) loggedInUsers(l *lua.LState) int {

	utmps, err := p.parseUtmpFile("/var/run/utmp")
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(utmps)
	return 1
}
