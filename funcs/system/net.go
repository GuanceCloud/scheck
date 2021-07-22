package system

import (
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"syscall"

	netutil "github.com/shirou/gopsutil/net"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system/impl"
)

func (p *provider) ipTables(l *lua.LState) int {

	data, err := ioutil.ReadFile(`/proc/net/ip_tables_names`)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable

	lines := strings.Split(string(data), "\n")
	for _, l := range lines {
		l = strings.TrimSpace(l)
		var item lua.LTable
		item.RawSetString("filter_name", lua.LString(l))
		result.Append(&item)
	}
	l.Push(&result)
	return 1
}

func (p *provider) processOpenSockets(l *lua.LState) int {

	var socketList []*impl.SocketInfo
	var err error

	var pids []int
	lv := l.Get(1)
	if lv != lua.LNil {
		if lv.Type() != lua.LTNumber {
			l.TypeError(1, lua.LTNumber)
			return lua.MultRet
		}
		pids = append(pids, int(lv.(lua.LNumber)))
	}

	socketList, err = impl.EnumProcessesOpenSockets(pids)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable
	for _, info := range socketList {
		var item lua.LTable
		item.RawSetString("socket", lua.LString(info.Socket))
		if info.Family == syscall.AF_UNIX {
			item.RawSetString("family", lua.LString("AF_UNIX"))
			item.RawSetString("protocol", lua.LString("ip"))
		} else {
			if info.Family == syscall.AF_INET {
				item.RawSetString("family", lua.LString("AF_INET"))
			} else if info.Family == syscall.AF_INET6 {
				item.RawSetString("family", lua.LString("AF_INET6"))
			}
			item.RawSetString("protocol", lua.LString(impl.LinuxProtocolNames[info.Protocol]))
		}

		item.RawSetString("local_address", lua.LString(info.LocalAddress))
		item.RawSetString("local_port", lua.LNumber(info.LocalPort))
		item.RawSetString("remote_address", lua.LString(info.RemoteAddress))
		item.RawSetString("remote_port", lua.LNumber(info.RemotePort))
		item.RawSetString("path", lua.LString(info.UnixSocketPath))
		item.RawSetString("state", lua.LString(info.State))
		item.RawSetString("net_namespace", lua.LNumber(info.Namespace))
		item.RawSetString("pid", lua.LNumber(info.PID))
		item.RawSetString("fd", lua.LNumber(info.Fd))

		if pname, _, pcmdline, err := impl.GetProcessSimpleInfo(info.PID); err == nil {
			item.RawSetString("process_name", lua.LString(pname))
			item.RawSetString("cmdline", lua.LString(pcmdline))
		}
		result.Append(&item)
	}

	l.Push(&result)
	return 1
}

func (p *provider)


listeningPorts(l *lua.LState) int {

	var pids []int
	lv := l.Get(1)
	if lv != lua.LNil {
		if lv.Type() != lua.LTNumber {
			l.TypeError(1, lua.LTNumber)
			return lua.MultRet
		}
		pids = append(pids, int(lv.(lua.LNumber)))
	}

	var socketList []*impl.SocketInfo
	var err error

	socketList, err = impl.EnumProcessesOpenSockets(pids)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable
	for _, info := range socketList {

		if info.Family == syscall.AF_UNIX && info.UnixSocketPath == "" {
			continue
		}

		if (info.Family == syscall.AF_INET || info.Family == syscall.AF_INET6) && info.RemotePort != 0 {
			continue
		}

		var item lua.LTable
		item.RawSetString("pid", lua.LNumber(info.PID))

		if pname, _, pcmdline, err := impl.GetProcessSimpleInfo(info.PID); err == nil {
			item.RawSetString("process_name", lua.LString(pname))
			item.RawSetString("cmdline", lua.LString(pcmdline))
		}

		if info.Family == syscall.AF_UNIX {
			item.RawSetString("port", lua.LNumber(0))
			item.RawSetString("path", lua.LString(info.UnixSocketPath))
			item.RawSetString("socket", lua.LString("0"))
			item.RawSetString("family", lua.LString("AF_UNIX"))
			item.RawSetString("protocol", lua.LString("ip"))
		} else {
			item.RawSetString("port", lua.LNumber(info.LocalPort))
			item.RawSetString("address", lua.LString(info.LocalAddress))
			item.RawSetString("socket", lua.LString(info.Socket))
			if info.Family == syscall.AF_INET {
				item.RawSetString("family", lua.LString("AF_INET"))
			} else if info.Family == syscall.AF_INET6 {
				item.RawSetString("family", lua.LString("AF_INET6"))
			}
			item.RawSetString("protocol", lua.LString(impl.LinuxProtocolNames[info.Protocol]))
		}
		item.RawSetString("fd", lua.LNumber(info.Fd))
		item.RawSetString("net_namespace", lua.LNumber(info.Namespace))
		result.Append(&item)
	}

	l.Push(&result)
	return 1
}

func (p *provider) interfaceAddresses(l *lua.LState) int {

	ifs, err := netutil.Interfaces()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	var result lua.LTable
	for _, it := range ifs {

		ip4 := ""
		ip6 := ""
		loopback := false
		for _, ad := range it.Addrs {
			ip, _, _ := net.ParseCIDR(ad.Addr)
			if ip.IsLoopback() {
				loopback = true
				continue
			}
			if ip.To4() != nil {
				ip4 = ad.Addr
			} else if ip.To16() != nil {
				ip6 = ad.Addr
			}
		}

		if loopback {
			continue
		}

		var eth lua.LTable
		eth.RawSetString("interface", lua.LString(it.Name))
		eth.RawSetString("ip4", lua.LString(ip4))
		eth.RawSetString("ip6", lua.LString(ip6))
		eth.RawSetString("mtu", lua.LNumber(it.MTU))
		eth.RawSetString("mac", lua.LString(it.HardwareAddr))
		result.Append(&eth)
	}

	l.Push(&result)
	return 1
}

func (p *provider) httpGet(l *lua.LState) int {

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	url := lv.(lua.LString)
	body, err := http.Get(url.String())
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	defer body.Body.Close()
	data, err := ioutil.ReadAll(body.Body)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(string(data)))
	return 1
}
