package net

import (
	"io/ioutil"
	"net"
	"net/http"

	netutil "github.com/shirou/gopsutil/net"
	lua "github.com/yuin/gopher-lua"
)

var apiLock = make(chan int, 1)
var api = map[string]lua.LGFunction{
	"interface_addresses": InterfaceAddresses,
	"http_get":            HTTPGet,
}

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

func InterfaceAddresses(l *lua.LState) int {
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

func HTTPGet(l *lua.LState) int {
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
