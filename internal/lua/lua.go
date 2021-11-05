// Package lua : func from github.com/yuin/gopher-lua
package lua

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"

	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
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

var supportLuaLibs = []luaLib{
	{lua.LoadLibName, lua.OpenPackage, nil},
	{lua.BaseLibName, lua.OpenBase, nil},
	{lua.TabLibName, lua.OpenTable, nil},
	{lua.StringLibName, lua.OpenString, nil},
	{lua.MathLibName, lua.OpenMath, nil},
	{lua.DebugLibName, lua.OpenDebug, nil},
	{lua.ChannelLibName, lua.OpenChannel, nil},
	{lua.OsLibName, lua.OpenOs, []string{"exit", "execute", "remove", "rename", "setenv", "setlocale"}},
}

var l = logger.DefaultSLogger("lua")

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
	LoadModule(ls)
	return &ScriptRunTime{Ls: ls}
}

func (r *ScriptRunTime) PCall(bcode *ByteCode) error {
	lfunc := r.Ls.NewFunctionFromProto(bcode.Proto)
	r.Ls.Push(lfunc)
	return r.Ls.PCall(0, lua.MultRet, nil)
}

func CompilesScript(filePath string) (*ByteCode, error) {
	file, err := os.Open(filepath.Clean(filePath))
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()
	reader := bufio.NewReader(file)
	chunk, err := luaparse.Parse(reader, filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to parse lua file '%s', err: %w", filePath, err)
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, fmt.Errorf("fail to compile lua file '%s', err: %w", filePath, err)
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

func LoadLuaLibs(ls *lua.LState) error {
	for _, lib := range supportLuaLibs {
		err := ls.CallByParam(lua.P{
			Fn:      ls.NewFunction(lib.fn),
			NRet:    1,
			Protect: true,
		}, lua.LString(lib.name))
		if err != nil {
			return fmt.Errorf("load %s failed, %w", lib.name, err)
		}
		lvMod := ls.Get(-1)
		if lvMod.Type() == lua.LTTable {
			lt, ok := lvMod.(*lua.LTable)
			if ok {
				for _, mth := range lib.disabledMethods {
					lt.RawSetString(mth, ls.NewClosure(unsupportFn, lua.LString(mth)))
				}
			}
		}
		ls.Pop(1)
	}
	return nil
}
