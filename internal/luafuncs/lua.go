package luafuncs

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"

	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
	luajson "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/json"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
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
		ID int
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
		return nil
	}
	luajson.Preload(ls)
	for _, p := range funcs.FuncProviders {
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
	t.RawSetString(global.LuaConfigurationKey, lua.LString(cfg.RulePath))
	l.SetGlobal(global.LuaConfiguration, &t)
}

func GetScriptGlobalConfig(l *lua.LState) *ScriptGlobalCfg {
	lv := l.GetGlobal(global.LuaConfiguration)
	if lv.Type() == lua.LTTable {
		t := lv.(*lua.LTable)
		var cfg ScriptGlobalCfg
		v := t.RawGetString(global.LuaConfigurationKey)
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
	luajson.Preload(ls) // for json parse
	for _, p := range funcs.FuncProviders {
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

func TestLua(rulepath string) {
	rulepath, _ = filepath.Abs(rulepath)
	if !strings.HasSuffix(rulepath, global.LuaExt) {
		rulepath += global.LuaExt
	}

	byteCode, err := CompilesScript(rulepath)
	if err != nil {
		fmt.Printf("Compile lua scripterr=%v \n", err)
		return
	}
	lua.LuaPathDefault = filepath.Join(global.InstallDir, global.DefRulesDir, "lib", "?.lua")

	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err = LoadLuaLibs(ls); err != nil {
		ls.Close()
		fmt.Printf("LoadLuaLibs err=%v \n ", err)
		return
	}
	SetScriptGlobalConfig(ls, &ScriptGlobalCfg{RulePath: rulepath})
	luajson.Preload(ls)
	for _, p := range funcs.FuncProviders {
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

// CheckLua check all custom lua.
func CheckLua(customRuleDir string) {
	fileInfos, err := ioutil.ReadDir(customRuleDir)
	if err != nil {
		l.Errorf("%v", err)
		return
	}
	if len(fileInfos) == 0 {
		fmt.Printf("there are no lua rules here %s \n", customRuleDir)
		return
	}
	errCount := 0
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}
		if strings.HasSuffix(info.Name(), ".lua") {
			_, err := CompilesScript(filepath.Join(customRuleDir, info.Name()))
			if err != nil {
				fmt.Printf("name of lua :%s compiles is err:%v \n", info.Name(), err)
				errCount++
			}
		}
		if strings.HasSuffix(info.Name(), ".manifest") {
			err := CompilesManifest(filepath.Join(customRuleDir, info.Name()))
			if err != nil {
				fmt.Printf("name of manifest :%s compiles is err:%v \n", info.Name(), err)
				errCount++
			}
		}
	}
	if errCount != 0 {
		fmt.Printf("there are %d error here \n", errCount)
	} else {
		fmt.Printf("all of the lua rules is ok! \n")
	}
}

func CompilesManifest(fileName string) error {
	var tbl *ast.Table
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	tbl, err = toml.Parse(contents)
	if err != nil {
		return err
	}
	requireKeys := map[string]bool{
		"id":       false,
		"category": false,
		"level":    false,
		"title":    false,
		"desc":     false,
		"cron":     false,
		"os_arch":  false,
	}
	for k := range requireKeys {
		v := tbl.Fields[k]
		if v == nil {
			continue
		}
		str := ""
		if kv, ok := v.(*ast.KeyValue); ok {
			if s, ok := kv.Value.(*ast.String); ok {
				str = s.Value
			}
			if s, ok := kv.Value.(*ast.Array); ok {
				str = s.Source()
			}
		}
		if str != "" {
			requireKeys[k] = true
		}
	}
	for name, ok := range requireKeys {
		if !ok {
			return fmt.Errorf("field name=%s can find", name)
		}
	}
	return nil
}
