package luaext

import (
	"io/ioutil"
	"net"
	"os"
	"strings"
	"syscall"

	netutil "github.com/shirou/gopsutil/net"
	lua "github.com/yuin/gopher-lua"
)

type (
	socketInfo struct {
		socket   string
		family   int
		protocol int
		ns       int64

		localAddress  string
		localPort     uint16
		remoteAddress string
		remotePort    uint16

		unixSocketPath string

		state string

		pid      int
		pname    string
		pcmdline string
		fd       string
	}
)

func hostname(l *lua.LState) int {
	name, err := os.Hostname()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(name))
	return 1
}

func ipTables(l *lua.LState) int {

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

func processOpenSockets(l *lua.LState) int {

	var socketList []*socketInfo
	var err error

	socketList, err = getProcessOpenSockets()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable
	for _, info := range socketList {

		var item lua.LTable
		item.RawSetString("socket", lua.LString(info.socket))
		if info.family == syscall.AF_UNIX {
			item.RawSetString("family", lua.LString("AF_UNIX"))
			item.RawSetString("protocol", lua.LString("ip"))
		} else {
			if info.family == syscall.AF_INET {
				item.RawSetString("family", lua.LString("AF_INET"))
			} else if info.family == syscall.AF_INET6 {
				item.RawSetString("family", lua.LString("AF_INET6"))
			}
			item.RawSetString("protocol", lua.LString(linuxProtocolNames[info.protocol]))
		}

		item.RawSetString("local_address", lua.LString(info.localAddress))
		item.RawSetString("local_port", lua.LNumber(info.localPort))
		item.RawSetString("remote_address", lua.LString(info.remoteAddress))
		item.RawSetString("remote_port", lua.LNumber(info.remotePort))
		item.RawSetString("path", lua.LString(info.unixSocketPath))
		item.RawSetString("state", lua.LString(info.state))
		item.RawSetString("net_namespace", lua.LNumber(info.ns))
		item.RawSetString("pid", lua.LNumber(info.pid))
		item.RawSetString("exe", lua.LString(info.pname))
		item.RawSetString("cmdline", lua.LString(info.pcmdline))
		item.RawSetString("fd", lua.LString(info.fd))
		result.Append(&item)
	}

	l.Push(&result)
	return 1
}

func listeningPorts(l *lua.LState) int {
	var socketList []*socketInfo
	var err error

	socketList, err = getProcessOpenSockets()
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	var result lua.LTable
	for _, info := range socketList {

		if info.family == syscall.AF_UNIX && info.unixSocketPath == "" {
			continue
		}

		if (info.family == syscall.AF_INET || info.family == syscall.AF_INET6) && info.remotePort != 0 {
			continue
		}

		var item lua.LTable
		item.RawSetString("pid", lua.LNumber(info.pid))
		item.RawSetString("exe", lua.LString(info.pname))
		item.RawSetString("cmdline", lua.LString(info.pcmdline))
		if info.family == syscall.AF_UNIX {
			item.RawSetString("port", lua.LNumber(0))
			item.RawSetString("path", lua.LString(info.unixSocketPath))
			item.RawSetString("socket", lua.LString("0"))
			item.RawSetString("family", lua.LString("AF_UNIX"))
			item.RawSetString("protocol", lua.LString("ip"))
		} else {
			item.RawSetString("port", lua.LNumber(info.localPort))
			item.RawSetString("address", lua.LString(info.localAddress))
			item.RawSetString("socket", lua.LString(info.socket))
			if info.family == syscall.AF_INET {
				item.RawSetString("family", lua.LString("AF_INET"))
			} else if info.family == syscall.AF_INET6 {
				item.RawSetString("family", lua.LString("AF_INET6"))
			}
			item.RawSetString("protocol", lua.LString(linuxProtocolNames[info.protocol]))
		}
		item.RawSetString("fd", lua.LString(info.fd))
		item.RawSetString("net_namespace", lua.LNumber(info.ns))
		result.Append(&item)
	}

	l.Push(&result)
	return 1
}

func interfaceAddresses(l *lua.LState) int {

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

func interfaceDetails(l *lua.LState) int {
	return 1
}
