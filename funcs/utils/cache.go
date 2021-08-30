package utils

import (
	"fmt"
	"sync"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
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
		t := lv.(*lua.LTable)
		t.ForEach(func(k lua.LValue, v lua.LValue) {
			newt.RawSet(k, v)
		})
		c.val = &newt
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
		t := c.val.(*lua.LTable)
		t.ForEach(func(k lua.LValue, v lua.LValue) {
			newt.RawSet(k, v)
		})
		return &newt
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
		log.Errorf("%s", err)
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

func (c *scriptCache) clean() {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.kvSorage = map[string]*cacheValue{}
	c.cur = 0
}

func (p *provider) setGlobalCache(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))

	lv = l.Get(2)
	if lv == lua.LNil {
		_ = globalCache.setKey(key, nil)
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

	val := &cacheValue{}
	val.fromLuaVal(lv)

	_ = globalCache.setKey(key, val)
	return 0
}

func (p *provider) getGlobalCache(l *lua.LState) int {

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return 0
	}
	key := string(lv.(lua.LString))

	val := globalCache.getKey(key)
	if val == nil {
		l.Push(lua.LNil)
	} else {
		l.Push(val.toLuaVal())
	}
	return 1
}

func (p *provider) setCache(l *lua.LState) int {

	globalCfg := luafuncs.GetScriptGlobalConfig(l)
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

	lv = l.Get(2)
	if lv == lua.LNil {
		_ = sc.setKey(key, nil)
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

	val := &cacheValue{}
	val.fromLuaVal(lv)

	_ = sc.setKey(key, val)
	return 0
}

func (p *provider) getCache(l *lua.LState) int {

	globalCfg := luafuncs.GetScriptGlobalConfig(l)
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

	val := sc.getKey(key)
	if val == nil {
		l.Push(lua.LNil)
	} else {
		l.Push(val.toLuaVal())
	}
	return 1
}
