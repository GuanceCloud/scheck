package luaext

import (
	"fmt"
	"io"
	"strings"

	luajson "github.com/layeh/gopher-json"
	lua "github.com/yuin/gopher-lua"
)

type (
	luaLib struct {
		name            string
		fn              lua.LGFunction
		disabledMethods []string
	}

	luaFunc struct {
		Name string
		Fn   lua.LGFunction
	}

	// LuaExt
	LuaExt struct {
	}
)

var (
	supportLuaLibs = []luaLib{
		{lua.BaseLibName, lua.OpenBase, nil},
		{lua.TabLibName, lua.OpenTable, nil},
		{lua.StringLibName, lua.OpenString, nil},
		{lua.MathLibName, lua.OpenMath, nil},
		{lua.DebugLibName, lua.OpenDebug, nil},
		{lua.OsLibName, lua.OpenOs, []string{"execute", "remove", "rename", "setenv", "setlocale"}},
	}

	luaExtendFuncs = []luaFunc{
		{`file_exist`, fileExist},
		{`file_info`, file_info},
		{`read_file`, readFile},
		{`file_hash`, fileHash},
		{`send_metric`, sendMetric},
		{`hostname`, hostname},
		{`uptime`, uptime},
		{`time_zone`, zone},
		{`kernel_info`, kernelInfo},
		{`kernel_modules`, kernelModules},
		{`mounts`, mounts},
		{`processes`, processes},
		{`interfaces`, netInterfaces},
		{`iptables`, ipTables},
		{`users`, users},
		{`shadow`, shadow},
		{`json_encode`, jsonEncode},
		{`json_decode`, jsonDecode},
	}
)

// NewLuaExt
func NewLuaExt() *LuaExt {
	l := &LuaExt{}
	return l
}

// Register extended lua funcs to lua machine
func (l *LuaExt) Register(lstate *lua.LState) error {
	if err := loadLuaLibs(lstate); err != nil {
		return err
	}
	luajson.Preload(lstate) //for json parse
	for _, f := range luaExtendFuncs {
		lstate.Register(f.Name, f.Fn)
	}
	return nil
}

func unsupportFn(ls *lua.LState) int {
	lv := ls.Get(lua.UpvalueIndex(1))
	ls.RaiseError("'%s' diabled", lv.String())
	return 0
}

func loadLuaLibs(ls *lua.LState) error {

	for _, lib := range supportLuaLibs {

		err := ls.CallByParam(lua.P{
			Fn:      ls.NewFunction(lib.fn),
			NRet:    1,
			Protect: true,
		}, lua.LString(lib.name))
		if err != nil {
			return fmt.Errorf("load %s failed, %s", lib.name, err)
		}
		lvMod := ls.Get(-1)
		if lvMod.Type() == lua.LTTable {
			lt := lvMod.(*lua.LTable)
			for _, mth := range lib.disabledMethods {
				lt.RawSetString(mth, ls.NewClosure(unsupportFn, lua.LString(mth)))
			}
		}
		ls.Pop(1)
	}
	return nil
}

func DumpSupportLuaFuncs(w io.Writer) {
	names := []string{}
	for idx, f := range luaExtendFuncs {
		names = append(names, fmt.Sprintf("  %d. %s", idx+1, f.Name))
	}
	s := strings.Join(names, "\n")
	s += "\n"
	w.Write([]byte(s))
}

func RunLuaScriptString(script string) error {
	ls := lua.NewState()
	defer ls.Close()
	le := NewLuaExt()
	if err := le.Register(ls); err != nil {
		return err
	}
	return ls.DoString(script)
}
