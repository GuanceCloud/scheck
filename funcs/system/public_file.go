package system

import (
	"bytes"
	"crypto/md5" // nolint:gosec
	"encoding/hex"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/gogf/gf/os/gfsnotify"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	l = logger.DefaultSLogger("func")
)

func (p *provider) ls(l *lua.LState) int {
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

	var files lua.LTable
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
			file := fileInfo2Table(f)
			file.RawSetString("path", lua.LString(filepath.Join(dir, f.Name())))
			files.Append(file)
		}
	} else {
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err == nil {
				if info != nil {
					file := fileInfo2Table(info)
					file.RawSetString("path", lua.LString(path))
					files.Append(file)
				}
			}
			return nil
		})
	}

	l.Push(&files)
	return 1
}

func (p *provider) fileExist(l *lua.LState) int {
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

func (p *provider) fileInfo(l *lua.LState) int {
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
	info := fileInfo2Table(stat)
	l.Push(info)
	return 1
}

func (p *provider) readFile(l *lua.LState) int {
	content := ""
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	content = string(data)
	l.Push(lua.LString(content))
	return 1
}

func (p *provider) fileHash(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	data, err := ioutil.ReadFile(path)
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

func (p *provider) grep(l *lua.LState) int {
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
		if errstr == "" {
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

func fileWatch(path string, scchan lua.LChannel) {
	go func() {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			l.Fatalf("func fileWatch error :%s", err)
			return
		}
		defer watcher.Close()

		done := make(chan bool, 1)
		go func() {
			for {
				select {
				case event, ok := <-watcher.Events:
					if !ok {
						return
					}
					var watch lua.LTable
					watch.RawSetString("path", lua.LString(path))
					watch.RawSetString("status", lua.LNumber(event.Op))
					scchan <- &watch
					if event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
						done <- true
					}
				case werr, ok := <-watcher.Errors:
					if !ok {
						return
					}
					l.Fatalf("func fileWatch error :%s", werr)
				}
			}
		}()
		err = watcher.Add(path)
		if err != nil {
			l.Fatalf("func fileWatch error :%s", err)
			return
		}
		<-done
	}()
}

func dirWatch(path string, scchan lua.LChannel) {
	go func() {
		done := make(chan bool, 1)

		_, err := gfsnotify.Add(path, func(event *gfsnotify.Event) {
			fmt.Println(event.Op)
			fmt.Println(event.Path)
			var watch lua.LTable
			watch.RawSetString("path", lua.LString(event.Path))
			watch.RawSetString("status", lua.LNumber(event.Op))
			scchan <- &watch
			//if event.IsRemove() || event.IsRename() {
			//	fi, err := os.Stat(path)
			//	if err == nil && !fi.IsDir() {
			//		done <- true
			//	}
			//}
		})
		if err != nil {
			l.Fatalf("func dirWatch error:%s ", err)
			return
		}
		<-done
	}()
}

func (p *provider) pathWatch(l *lua.LState) int {
	var strN = 1
	var chanN = 2
	path := l.ToString(strN)
	scchan := l.ToChannel(chanN)
	_, err := os.Stat(path)
	if err != nil {
		return lua.MultRet
	} else {
		dirWatch(path, scchan)
	}

	return 0
}
