package funcs

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
	checker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	luajson "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/json"
)

type (
	luaLib struct {
		name            string
		fn              lua.LGFunction
		disabledMethods []string
	}

	ByteCode struct {
		Proto *lua.FunctionProto
	}

	ScriptRunTime struct {
		Id int
		Ls *lua.LState
	}
)

var (
	supportLuaLibs = []luaLib{
		{lua.LoadLibName, lua.OpenPackage, nil},
		{lua.BaseLibName, lua.OpenBase, nil},
		{lua.TabLibName, lua.OpenTable, nil},
		{lua.StringLibName, lua.OpenString, nil},
		{lua.MathLibName, lua.OpenMath, nil},
		{lua.DebugLibName, lua.OpenDebug, nil},
		{lua.ChannelLibName, lua.OpenChannel, nil},
		{lua.OsLibName, lua.OpenOs, []string{"exit", "execute", "remove", "rename", "setenv", "setlocale"}},
	}
)

func (r *ScriptRunTime) Close() {
	if r.Ls != nil {
		if !r.Ls.IsClosed() {
			r.Ls.Close()
		}
	}
}

func NewScriptRunTime() *ScriptRunTime {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := LoadLuaLibs(ls); err != nil {
		ls.Close()
		//l.Errorf("LoadLuaLibs err=%v \n ", err)
		return nil
	}
	luajson.Preload(ls)
	for _, p := range checker.FuncProviders {
		for _, f := range p.Funcs() {
			ls.Register(f.Name, lua.LGFunction(f.Fn))
		}
	}
	return &ScriptRunTime{Ls: ls}
}

func (r *ScriptRunTime) PCall(bcode *ByteCode) error {
	lfunc := r.Ls.NewFunctionFromProto(bcode.Proto)
	r.Ls.Push(lfunc)
	return r.Ls.PCall(0, lua.MultRet, nil)
}

func CompilesScript(filePath string) (*ByteCode, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	chunk, err := luaparse.Parse(reader, filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to parse lua file '%s', err: %s", filePath, err)
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to compile lua file '%s', err: %s", filePath, err)
	}
	return &ByteCode{
		Proto: proto,
	}, nil
}

func unsupportFn(ls *lua.LState) int {
	lv := ls.Get(lua.UpvalueIndex(1))
	ls.RaiseError("'%s' diabled", lv.String())
	return 0
}

type ScriptGlobalCfg struct {
	RulePath string
}

func SetScriptGlobalConfig(l *lua.LState, cfg *ScriptGlobalCfg) {
	var t lua.LTable
	t.RawSetString("rulefile", lua.LString(cfg.RulePath))
	l.SetGlobal("__this_configuration", &t)
}

func GetScriptGlobalConfig(l *lua.LState) *ScriptGlobalCfg {
	lv := l.GetGlobal("__this_configuration")
	if lv.Type() == lua.LTTable {
		t := lv.(*lua.LTable)
		var cfg ScriptGlobalCfg
		v := t.RawGetString("rulefile")
		if v.Type() == lua.LTString {
			cfg.RulePath = string(v.(lua.LString))
		}
		return &cfg
	}
	return nil
}

func GetScriptRuntime(cfg *ScriptGlobalCfg) (*ScriptRunTime, error) {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := LoadLuaLibs(ls); err != nil {
		ls.Close()
		return nil, err
	}
	SetScriptGlobalConfig(ls, cfg)
	luajson.Preload(ls) //for json parse
	for _, p := range checker.FuncProviders {
		for _, f := range p.Funcs() {
			ls.Register(f.Name, lua.LGFunction(f.Fn))
		}
	}
	return &ScriptRunTime{
		Ls: ls,
	}, nil
}

func LoadLuaLibs(ls *lua.LState) error {

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

// testLua
func TestLua(rulepath string) {
	rulepath, _ = filepath.Abs(rulepath)
	rulepath = rulepath + ".lua"
	byteCode, err := CompilesScript(rulepath)
	if err != nil {
		fmt.Printf("Compile lua scripterr=%v \n", err)
		return
	}
	lua.LuaPathDefault = "./rules.d/lib/?.lua"

	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err := LoadLuaLibs(ls); err != nil {
		ls.Close()
		fmt.Printf("LoadLuaLibs err=%v \n ", err)
		return
	}
	SetScriptGlobalConfig(ls, &ScriptGlobalCfg{RulePath: rulepath})
	luajson.Preload(ls)
	for _, p := range checker.FuncProviders {
		for _, f := range p.Funcs() {
			ls.Register(f.Name, lua.LGFunction(f.Fn))
		}
	}

	lfunc := ls.NewFunctionFromProto(byteCode.Proto)
	ls.Push(lfunc)
	if err = ls.PCall(0, lua.MultRet, nil); err != nil {
		fmt.Printf("testLua err=%v \n", err)
	}

}
