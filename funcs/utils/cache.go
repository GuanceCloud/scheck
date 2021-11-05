// Package utils has json,mysql,cache,utils and so on.
package utils

import (
	"fmt"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

const (
	MaxCacheSizePerScript = 1024 * 1024
)

type (
	cacheValue struct {
		typ lua.LValueType
		val interface{}
		len int
	}

	scriptCache struct {
		cur      int
		kvSorage map[string]*cacheValue
		mux      sync.Mutex
	}

	cachePool struct {
		scripts map[string]*scriptCache
		mux     sync.Mutex
	}
)

// ScriptGlobalCfg :lua运行时回调go函数时，go无法得知是哪个脚本，
// 所以在每个脚本运行之前都放到global中一个键值对(key:文件名).
type ScriptGlobalCfg struct {
	RulePath string
}

var cacheDB = cachePool{
	scripts: map[string]*scriptCache{},
}

var globalCache = &scriptCache{
	kvSorage: map[string]*cacheValue{},
}

func (c *cacheValue) fromLuaVal(lv lua.LValue) {
	c.typ = lv.Type()
	switch c.typ {
	case lua.LTBool:
		c.val = bool(lv.(lua.LBool))
		c.len = 1
	case lua.LTNumber:
		c.val = float64(lv.(lua.LNumber))
		c.len = 8
	case lua.LTString:
		c.val = lv.String()
		c.len = len(lv.String())
	case lua.LTTable:
		var newt lua.LTable
		t, ok := lv.(*lua.LTable)
		if ok {
			t.ForEach(func(k lua.LValue, v lua.LValue) {
				newt.RawSet(k, v)
			})
		}
		c.val = &newt
	case lua.LTChannel, lua.LTFunction, lua.LTNil, lua.LTThread, lua.LTUserData:
	default:
	}
}

func (c *cacheValue) toLuaVal() lua.LValue {
	switch c.typ {
	case lua.LTBool:
		return lua.LBool(c.val.(bool))
	case lua.LTNumber:
		return lua.LNumber(c.val.(float64))
	case lua.LTString:
		return lua.LString(c.val.(string))
	case lua.LTTable:
		var newt lua.LTable
		t, ok := c.val.(*lua.LTable)
		if ok {
			t.ForEach(func(k lua.LValue, v lua.LValue) {
				newt.RawSet(k, v)
			})
		}
		return &newt
	case lua.LTChannel, lua.LTFunction, lua.LTNil, lua.LTThread, lua.LTUserData:
	default:
	}
	return lua.LNil
}

func (p *cachePool) getScriptCache(script string) *scriptCache {
	p.mux.Lock()
	defer p.mux.Unlock()
	c := cacheDB.scripts[script]
	if c == nil {
		c = &scriptCache{
			kvSorage: map[string]*cacheValue{},
		}
		cacheDB.scripts[script] = c
	}
	return c
}

func (c *scriptCache) setKey(key string, val *cacheValue) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.cur+val.len > MaxCacheSizePerScript {
		err := fmt.Errorf("cache size limit exceeded")
		return err
	}
	c.kvSorage[key] = val
	c.cur += val.len
	return nil
}

func (c *scriptCache) getKey(key string) *cacheValue {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.kvSorage[key]
}

// nolint:unused
func (c *scriptCache) clean() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.kvSorage = map[string]*cacheValue{}
	c.cur = 0
}

func (c *scriptCache) cleanKey(key string) {
	c.mux.Lock()
	defer c.mux.Unlock()
	clen := c.kvSorage[key].len
	delete(c.kvSorage, key)
	c.cur -= clen
}

func SetGlobalCache(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))

	lv = l.Get(global.LuaArgIdx2)
	if lv == lua.LNil {
		_ = globalCache.setKey(key, nil)
		return 0
	}
	switch lv.Type() {
	case lua.LTBool:
	case lua.LTNumber:
	case lua.LTString:
	case lua.LTTable:
	case lua.LTChannel, lua.LTFunction, lua.LTNil, lua.LTThread, lua.LTUserData:
		l.RaiseError("invalid value type %s, only support boolean', 'string', 'number'", lv.Type().String())
		return 0
	default:
	}

	val := &cacheValue{}
	val.fromLuaVal(lv)

	_ = globalCache.setKey(key, val)
	return 0
}

func GetGlobalCache(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))

	if val := globalCache.getKey(key); val == nil {
		l.Push(lua.LNil)
	} else {
		l.Push(val.toLuaVal())
	}
	return 1
}

func SetCache(l *lua.LState) int {
	globalCfg := GetScriptGlobalConfig(l)
	if globalCfg == nil || globalCfg.RulePath == "" {
		return 0
	}
	sc := cacheDB.getScriptCache(globalCfg.RulePath)
	if sc == nil {
		return 0
	}
	lv := l.Get(global.LuaArgIdx1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))
	lv = l.Get(global.LuaArgIdx2)
	if lv == lua.LNil {
		_ = sc.setKey(key, nil)
		return 0
	}
	switch lv.Type() {
	case lua.LTBool:
	case lua.LTNumber:
	case lua.LTString:
	case lua.LTTable:
	case lua.LTChannel, lua.LTFunction, lua.LTNil, lua.LTThread, lua.LTUserData:
		l.RaiseError("invalid value type %s, only support boolean', 'string', 'number'", lv.Type().String())
		return 0
	}

	val := &cacheValue{}
	val.fromLuaVal(lv)
	_ = sc.setKey(key, val)
	return 0
}

func GetCache(l *lua.LState) int {
	globalCfg := GetScriptGlobalConfig(l)
	if globalCfg == nil || globalCfg.RulePath == "" {
		return 0
	}
	sc := cacheDB.getScriptCache(globalCfg.RulePath)
	if sc == nil {
		return 0
	}
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))
	if val := sc.getKey(key); val == nil {
		l.Push(lua.LNil)
	} else {
		l.Push(val.toLuaVal())
	}
	return 1
}

func CleanCache(l *lua.LState) int {
	globalCfg := GetScriptGlobalConfig(l)
	if globalCfg == nil || globalCfg.RulePath == "" {
		return 0
	}
	sc := cacheDB.getScriptCache(globalCfg.RulePath)
	if sc == nil {
		return 0
	}
	sc.clean()
	return 0
}

func DeleteCache(l *lua.LState) int {
	globalCfg := GetScriptGlobalConfig(l)
	if globalCfg == nil || globalCfg.RulePath == "" {
		return 0
	}
	sc := cacheDB.getScriptCache(globalCfg.RulePath)
	if sc == nil {
		return 0
	}

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))
	sc.cleanKey(key)
	return 0
}

func DeleteCacheAll(l *lua.LState) int {
	// delete all rules cache
	cacheDB.mux.Lock()
	cacheDB.scripts = make(map[string]*scriptCache)
	cacheDB.mux.Unlock()
	return 0
}

func SetScriptGlobalConfig(l *lua.LState, cfg *ScriptGlobalCfg) {
	var t lua.LTable
	t.RawSetString(global.LuaConfigurationKey, lua.LString(cfg.RulePath))
	l.SetGlobal(global.LuaConfiguration, &t)
}

func GetScriptGlobalConfig(l *lua.LState) *ScriptGlobalCfg {
	lv := l.GetGlobal(global.LuaConfiguration)
	if lv.Type() == lua.LTTable {
		t, ok := lv.(*lua.LTable)
		if ok {
			var cfg ScriptGlobalCfg
			v := t.RawGetString(global.LuaConfigurationKey)
			if v.Type() == lua.LTString {
				cfg.RulePath = string(v.(lua.LString))
			}
			return &cfg
		}
	}
	return nil
}
