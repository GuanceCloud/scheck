package luafuncs

import (
	"strconv"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

// nolint
func BenchmarkScriptRunTime_PCall_1(b *testing.B) {
	name := "./examples.lua"
	state := &ScriptRunTime{Ls: lua.NewState()}
	bCode, err := CompilesScript(name)
	if err != nil {
		b.Fatalf("Compiles Script name = %s ,err=%v", name, err)
	}
	for i := 0; i < b.N; i++ {
		state.Ls.Register("put_to_table", lua.LGFunction(putToTable))

		lFunc := state.Ls.NewFunctionFromProto(bCode.Proto)
		state.Ls.Push(lFunc)

		if err = state.Ls.PCall(0, lua.MultRet, nil); err != nil {
			b.Error(err)
		}
	}
}

// nolint
// 每次都是初始化新的state
func BenchmarkScriptRunTime_PCall_2(b *testing.B) {
	name := "./examples.lua"
	bCode, err := CompilesScript(name)
	if err != nil {
		b.Fatalf("Compiles Script name = %s ,err=%v", name, err)
	}

	for i := 0; i < b.N; i++ {
		state := &ScriptRunTime{Ls: lua.NewState()}
		state.Ls.Register("put_to_table", lua.LGFunction(putToTable))

		lFunc := state.Ls.NewFunctionFromProto(bCode.Proto)
		state.Ls.Push(lFunc)

		if err = state.Ls.PCall(0, lua.MultRet, nil); err != nil {
			b.Error(err)
		}
	}
}

// nolint
func putToTable(ls *lua.LState) int {
	var files lua.LTable
	for i := 0; i < 100; i++ {
		var file lua.LTable
		key := "x" + strconv.Itoa(i) // key 不可以是纯数字的字符串  val:长字符串
		val := "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system.fileInfo2Table in /home/gopath/src/gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system/file_linux.gogitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system.fileInfo2Table in /home/gopath/src/gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system/file_linux.go"
		file.RawSetString(key, lua.LString(val))
		files.Append(&file)
	}
	ls.Push(&files)
	return 1
}
