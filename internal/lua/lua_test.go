package lua

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	lua "github.com/yuin/gopher-lua"
)

var testFile = "./testdata"

// nolint
func BenchmarkScriptRunTime_PCall_1(b *testing.B) {
	name := filepath.Join(testFile, "examples.lua")
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
	name := filepath.Join(testFile, "examples.lua")
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

func BenchmarkLoadLuaLibs(b *testing.B) {
	runtimeSc := NewScriptRunTime()
	var err error
	for i := 0; i < b.N; i++ {
		err = LoadLuaLibs(runtimeSc.Ls)
		if err != nil {
			b.Logf("err=%v", err)
			return
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

func TestCompilesManifest(t *testing.T) {
	// load from file or read from string
	var err error
	var contents []byte
	var tbl *ast.Table
	contents, err = ioutil.ReadFile("testdata/0001-test.manifest")
	if err != nil {
		l.Warnf("read file err=%v", err)
		return
	}
	// 去掉有可能在UTF8编码中存在的BOM头
	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))
	tbl, err = toml.Parse(contents)
	if err != nil {
		l.Warnf("toml.Parse err=%v", err)
		return
	}
	for k, v := range tbl.Fields {
		t.Log(k)
		switch table := v.(type) {
		case *ast.String:
			t.Log("是string类型")
		case *ast.Array:
			t.Log("是数组类型")
		case *ast.KeyValue:
			t.Log("is KeyValue")
		case *ast.Table:
			table.Source()
			t.Logf("%+v", table.Fields)
		default:
			t.Log("Unknown type : v")
		}
	}
}

func TestCompilesScript(t *testing.T) {
	_, err := CompilesScript(filepath.Join(testFile, "hostname.lua"))
	if err != nil {
		t.Logf("CompilesScript err =%v", err)
		return
	}
}

func TestNewScriptRunTime(t *testing.T) {
	lua.LuaPathDefault = "./testdata/lib/?.lua"
	InitModules()                   // 1.先初始化module
	runtimeSc := NewScriptRunTime() // 2.初始化lua.state
	LoadModule(runtimeSc.Ls)        // 3.将模块加载进去
	btscode, err := CompilesScript(filepath.Join(testFile, "hostname.lua"))
	if err != nil {
		t.Logf("CompilesScript err =%v", err)
		return
	}

	fn := runtimeSc.Ls.NewFunctionFromProto(btscode.Proto)
	runtimeSc.Ls.Push(fn)
	if err = runtimeSc.Ls.PCall(0, lua.MultRet, nil); err != nil {
		t.Logf("pcall err =%v", err)
		return
	}
	t.Log("run lua proto ok")
}
