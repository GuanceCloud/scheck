package utils

import (
	luajson "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
)

func (p *provider) jsonEncode(l *lua.LState) int {

	lv := l.Get(1)
	data, err := luajson.Encode(lv)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(lua.LString(string(data)))
	return 1
}

func (p *provider) jsonDecode(l *lua.LState) int {

	lv := l.Get(1)
	if v, ok := lv.(lua.LString); !ok {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	} else {
		delv, err := luajson.Decode(l, []byte(v))
		if err != nil {
			l.RaiseError("%s", err)
			return lua.MultRet
		}
		l.Push(delv)
	}

	return 1
}
