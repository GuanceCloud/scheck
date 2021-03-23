package luaext

import (
	"fmt"

	luajson "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
)

func table2map(tbl *lua.LTable) (map[string]interface{}, error) {
	fields := map[string]interface{}{}

	var err error

	cb := func(k lua.LValue, v lua.LValue) {
		if err != nil {
			return
		}

		switch v.Type() {
		case lua.LTBool:
		case lua.LTNumber:
		case lua.LTString:
		case lua.LTTable:
		default:
			err = fmt.Errorf("invalid value of fields key: '%s'. fields value only support 'boolean', 'string', 'number', 'table'", k.String())
		}
		if err == nil {
			var val interface{}
			if v.Type() == lua.LTTable {
				if st, e := table2map(v.(*lua.LTable)); e != nil {
					err = e
					return
				} else {
					val = st
				}
			} else {
				if v.Type() == lua.LTNumber {
					num := v.(lua.LNumber)
					if float64(num) == float64(int64(num)) {
						val = int64(num)
					} else {
						val = float64(num)
					}
				} else {
					val = v.String()
				}
			}

			fields[k.String()] = val
		}
	}

	tbl.ForEach(cb)

	if err != nil {
		return nil, err
	}
	return fields, nil
}

func jsonEncode(l *lua.LState) int {

	lv := l.Get(1)
	data, err := luajson.Encode(lv)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	l.Push(lua.LString(string(data)))
	return 1
}

func jsonDecode(l *lua.LState) int {

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
