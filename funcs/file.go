package funcs

import (
	"io/ioutil"
	"os"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

func fileExist(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	exist := false
	if path != "" {
		if _, err := os.Stat(path); err == nil {
			exist = true
		}
	}
	l.Push(lua.LBool(exist))
	return 1
}

func readFile(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(string(data)))
	return 1
}
