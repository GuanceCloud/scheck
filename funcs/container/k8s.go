// Package container has docker and k8s func
package container

import (
	"strconv"
	"strings"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/container/impl/utils"
)

func TLSCipherSuites(l *lua.LState) int {
	// 验证tls-cipher-suites 加密类型的方式
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	str := string(lv.(lua.LString))
	cmds := strings.Split(str, ",")
	if len(cmds) == 0 {
		l.Push(lua.LFalse)
		return 1
	}
	flag := true
	for _, cmd := range cmds {
		switch cmd {
		case "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256":
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
		case "TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305":
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
		case "TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305":
		case "TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384":
		case "TLS_RSA_WITH_AES_256_GCM_SHA384":
		case "TLS_RSA_WITH_AES_128_GCM_SHA256":
		default:
			flag = false
		}
	}
	if !flag {
		l.Push(lua.LFalse)
		return 1
	}
	l.Push(lua.LTrue)
	return 1
}

// nolint
func CheckVersion(l *lua.LState) int {
	// 当前规则 最低版本是 1.16
	mor := 1
	min := 16
	var v1, v2 int
	var strs []string
	version, err := utils.GetKubectlVersion()
	if err != nil {
		goto Error
	}
	strs = strings.Split(version, ".")
	if len(strs) < 2 {
		goto Error
	}
	v1, err = strconv.Atoi(strs[0])
	v2, err = strconv.Atoi(strs[1])
	if err != nil {
		goto Error
	}
	if v1 == mor {
		if v2 >= min {
			l.Push(lua.LTrue)
			return 1
		}
	}
Error:
	l.Push(lua.LFalse)
	return 1
}
