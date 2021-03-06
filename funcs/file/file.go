// Package file is go export for lua
package file

import (
	"bytes"

	// nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/os/gfsnotify"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

var l = logger.DefaultSLogger("func")

var api = map[string]lua.LGFunction{
	"ls":            Ls,
	"file_exist":    Exist,
	"file_info":     Info,
	"read_file":     ReadFile,
	"file_hash":     Hash,
	"grep":          Grep,
	"sc_path_watch": PathWatch,
}

func Loader(l *lua.LState) int {
	t := l.NewTable()
	mod := l.SetFuncs(t, api)
	l.Push(mod)
	return 1
}

func Ls(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	dir := string(lv.(lua.LString))
	dir = strings.TrimSpace(dir)

	rescue := false
	lv = l.Get(global.LuaArgIdx2)
	if lv != lua.LNil {
		if lv.Type() != lua.LTBool {
			l.TypeError(1, lua.LTBool)
			return lua.MultRet
		}
		rescue = bool(lv.(lua.LBool))
	}

	files := l.NewTable()

	if !rescue {
		list, err := ioutil.ReadDir(dir)
		if err != nil {
			l.RaiseError("%s", err)
			return lua.MultRet
		}
		for _, f := range list {
			if f == nil {
				continue
			}
			file := l.NewTable()
			fileInfo2Table(f, file)
			file.RawSetString("path", lua.LString(filepath.Join(dir, f.Name())))
			files.Append(file)
		}
	} else {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if info != nil {
					file := l.NewTable()
					fileInfo2Table(info, file)
					file.RawSetString("path", lua.LString(path))
					files.Append(file)
				}
			}
			return nil
		})
	}

	l.Push(files)
	return 1
}

func Exist(l *lua.LState) int {
	lv := l.Get(1)
	exist := false
	if lv.Type() == lua.LTString {
		path := string(lv.(lua.LString))
		path = strings.TrimSpace(path)
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				exist = true
			}
		}
	}

	l.Push(lua.LBool(exist))
	return 1
}

func Info(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	stat, err := os.Stat(path)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	info := l.NewTable()
	fileInfo2Table(stat, info)
	l.Push(info)
	return 1
}

func ReadFile(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	l.Push(lua.LString(string(data)))
	return 1
}

func Hash(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	data, err := ioutil.ReadFile(filepath.Clean(path))
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	m := md5.New() // nolint:gosec
	m.Write(data)
	result := hex.EncodeToString(m.Sum(nil))

	l.Push(lua.LString(result))
	return 1
}

func Grep(l *lua.LState) int {
	var opts string
	lv := l.Get(1)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		}
		opts = lv.(lua.LString).String()
	}

	var pattern string
	lv = l.Get(global.LuaArgIdx2)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		}
		pattern = lv.(lua.LString).String()
	}

	var filearg string
	lv = l.Get(global.LuaArgIdx3)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		}
		filearg = lv.(lua.LString).String()
	}

	cmd := exec.Command("grep")
	if opts != "" {
		cmd.Args = append(cmd.Args, opts)
	}
	if pattern != "" {
		cmd.Args = append(cmd.Args, pattern)
	}
	if filearg != "" {
		cmd.Args = append(cmd.Args, filearg)
	}

	buf := bytes.NewBuffer([]byte{})
	errbuf := bytes.NewBuffer([]byte{})
	cmd.Stdout = buf
	cmd.Stderr = errbuf
	err := cmd.Run()
	if cmd.ProcessState.ExitCode() != 0 {
		if cmd.ProcessState.ExitCode() == 1 {
			l.Push(lua.LString(""))
			l.Push(lua.LString(""))
			return global.LuaRet2
		}
		errstr := errbuf.String()
		if errstr == "" && err != nil {
			errstr = err.Error()
		}

		l.Push(lua.LString(""))
		l.Push(lua.LString(errstr))
		return global.LuaRet2
	}
	l.Push(lua.LString(buf.String()))
	l.Push(lua.LString(""))
	return global.LuaRet2
}

func dirWatch(path string, scchan lua.LChannel) {
	go func() {
		done := make(chan bool, 1)

		_, err := gfsnotify.Add(path, func(event *gfsnotify.Event) {
			var watch lua.LTable
			watch.RawSetString("path", lua.LString(event.Path))
			watch.RawSetString("status", lua.LNumber(event.Op))
			scchan <- &watch
		})
		if err != nil {
			l.Fatalf("func dirWatch error:%s ", err)
			return
		}
		<-done
	}()
}

func PathWatch(l *lua.LState) int {
	strN := 1
	chanN := 2
	path := l.ToString(strN)
	scchan := l.ToChannel(chanN)
	if _, err := os.Stat(path); err != nil {
		return lua.MultRet
	}
	dirWatch(path, scchan)

	return 0
}
