package luaext

import (
	"io/ioutil"
	"net"
	"os"
	"strings"

	netutil "github.com/shirou/gopsutil/net"
	lua "github.com/yuin/gopher-lua"
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

func netInterfaces(l *lua.LState) int {

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

func ports(l *lua.LState) int {
	return 1
}

func routes(l *lua.LState) int {
	return 1
}

func ipInfo(l *lua.LState) int {
	return 1
}
