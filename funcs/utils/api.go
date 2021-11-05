package utils

import (
	lua "github.com/yuin/gopher-lua"
)

var jsonAPI = map[string]lua.LGFunction{
	"json_encode": JSONEncode,
	"json_decode": JSONDecode,
}

var cacheAPI = map[string]lua.LGFunction{
	"set_cache":        SetCache,
	"get_cache":        GetCache,
	"del_cache":        DeleteCache,
	"clean_cache":      CleanCache,
	"del_cache_all":    DeleteCacheAll,
	"set_global_cache": SetGlobalCache,
	"get_global_cache": GetGlobalCache,
}

var mysqlAPI = map[string]lua.LGFunction{
	"mysql_weak_psw":   CheckMysqlWeakPassword,
	"mysql_ports_list": MysqlPortsList,
}

var utilsAPI = map[string]lua.LGFunction{
	"get_command_value": CommandValue,
}

func UtilsLoader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, utilsAPI)
	l.Push(mod)
	return 1
}

func JSONLoader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, jsonAPI)
	l.Push(mod)
	return 1
}

func CacheLoader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, cacheAPI)
	l.Push(mod)
	return 1
}

func MysqlLoader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, mysqlAPI)
	l.Push(mod)
	return 1
}
