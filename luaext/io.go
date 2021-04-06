package luaext

import (
	"sync"

	lua "github.com/yuin/gopher-lua"
)

const OneFileGlobalConfKey = "__this_configuration"

func setCache(l *lua.LState) int {

	filename := ""
	lv := l.GetGlobal(OneFileGlobalConfKey)
	if lv.Type() == lua.LTTable {
		t := lv.(*lua.LTable)
		rulefile := t.RawGetString("rulefile")
		if rulefile != lua.LNil && rulefile.Type() == lua.LTString {
			filename = string(rulefile.(lua.LString))
		}
	}

	if filename == "" {
		return 0
	}

	var key string

	lv = l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key = string(lv.(lua.LString))

	lv = l.Get(2)
	if lv == lua.LNil {
		setKey(filename, key, nil)
		return 0
	}
	switch lv.Type() {
	case lua.LTBool:
	case lua.LTNumber:
	case lua.LTString:
	case lua.LTTable:

	default:
		l.RaiseError("invalid value type %s, only support boolean', 'string', 'number'", lv.Type().String())
		return 0
	}

	val := &luaCacheValue{}
	val.fromLuaVal(lv)

	setKey(filename, key, val)
	return 0
}

func getCache(l *lua.LState) int {

	filename := ""
	lv := l.GetGlobal(OneFileGlobalConfKey)
	if lv.Type() == lua.LTTable {
		t := lv.(*lua.LTable)
		lvfilename := t.RawGetString("rulefile")
		if lvfilename != lua.LNil && lvfilename.Type() == lua.LTString {
			filename = string(lvfilename.(lua.LString))
		}
	}

	if filename == "" {
		l.Push(lua.LNil)
		return 1
	}

	var key string

	lv = l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key = string(lv.(lua.LString))

	val := getKey(filename, key)
	if val == nil {
		l.Push(lua.LNil)
	} else {
		l.Push(val.toLuaVal())
	}
	return 1
}

type luaKVCache map[string]*luaCacheValue

type luaCacheValue struct {
	typ lua.LValueType
	val interface{}
}

func (c *luaCacheValue) fromLuaVal(lv lua.LValue) {
	c.typ = lv.Type()
	switch c.typ {
	case lua.LTBool:
		c.val = lv.(lua.LBool)
	case lua.LTNumber:
		c.val = float64(lv.(lua.LNumber))
	case lua.LTString:
		c.val = lv.String()
	case lua.LTTable:
		var newt lua.LTable
		t := lv.(*lua.LTable)
		t.ForEach(func(k lua.LValue, v lua.LValue) {
			newt.RawSet(k, v)
		})
		c.val = &newt
	}
}

func (c *luaCacheValue) toLuaVal() lua.LValue {
	switch c.typ {
	case lua.LTBool:
		return lua.LBool(c.val.(bool))
	case lua.LTNumber:
		return lua.LNumber(c.val.(float64))
	case lua.LTString:
		return lua.LString(c.val.(string))
	case lua.LTTable:
		var newt lua.LTable
		t := c.val.(*lua.LTable)
		t.ForEach(func(k lua.LValue, v lua.LValue) {
			newt.RawSet(k, v)
		})
		return &newt
	}
	return lua.LNil
}

var (
	cacheDB  map[string]luaKVCache = map[string]luaKVCache{}
	cacheMux sync.Mutex
)

func setKey(filename string, key string, val *luaCacheValue) {
	cacheMux.Lock()
	defer cacheMux.Unlock()
	c := cacheDB[filename]
	if c == nil {
		cacheDB[filename] = make(luaKVCache)
		c = cacheDB[filename]
	}
	c[key] = val
}

func getKey(filename string, key string) *luaCacheValue {
	cacheMux.Lock()
	defer cacheMux.Unlock()
	c := cacheDB[filename]
	if c == nil {
		return nil
	}
	return c[key]
}
