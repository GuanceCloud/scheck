package luaext

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	lua "github.com/yuin/gopher-lua"
)

func fileExist(l *lua.LState) int {
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

func file_info(l *lua.LState) int {

	var info lua.LTable

	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		errstr := fmt.Sprintf("bad argument 1 (%v expected, got %v)", lua.LTString.String(), lv.Type().String())
		l.Push(lua.LNil)
		l.Push(lua.LString(errstr))
		return 2
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	stat, err := os.Stat(path)
	if err != nil {
		l.Push(lua.LNil)
		l.Push(lua.LString(err.Error()))
		return 2
	}

	st := stat.Sys().(*syscall.Stat_t)

	info.RawSetString("size", lua.LNumber(stat.Size()))
	info.RawSetString("block_size", lua.LNumber(st.Blksize))
	info.RawSetString("mode", lua.LNumber(st.Mode))
	info.RawSetString("uid", lua.LNumber(st.Uid))
	info.RawSetString("gid", lua.LNumber(st.Gid))
	info.RawSetString("device", lua.LNumber(st.Dev))
	info.RawSetString("inode", lua.LNumber(st.Ino))
	info.RawSetString("hard_links", lua.LNumber(st.Nlink))

	info.RawSetString("ctime", lua.LNumber(st.Ctim.Sec))
	info.RawSetString("mtime", lua.LNumber(st.Mtim.Sec))
	info.RawSetString("atime", lua.LNumber(st.Atim.Sec))

	l.Push(&info)
	l.Push(lua.LString(""))
	return 2
}

func readFile(l *lua.LState) int {
	content := ""
	errstr := ""
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		errstr = fmt.Sprintf("bad argument 1 (%v expected, got %v)", lua.LTString.String(), lv.Type().String())
		l.Push(lua.LString(content))
		l.Push(lua.LString(errstr))
		return 2
	}

	path := string(lv.(lua.LString))
	path = strings.TrimSpace(path)
	if data, err := ioutil.ReadFile(path); err != nil {
		errstr = err.Error()
	} else {
		content = string(data)
	}
	l.Push(lua.LString(content))
	l.Push(lua.LString(errstr))
	return 2
}

func fileHash(l *lua.LState) int {

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

	m := md5.New()
	m.Write(data)
	result := hex.EncodeToString(m.Sum(nil))

	l.Push(lua.LString(result))
	return 1
}
