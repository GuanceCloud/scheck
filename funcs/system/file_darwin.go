//+build darwin
package system

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	lua "github.com/yuin/gopher-lua"
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
	lv = l.Get(2)
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
		filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {

			}
			if info != nil {
				file := fileInfo2Table(info)
				file.RawSetString("path", lua.LString(path))
				files.Append(file)
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
	if data, err := ioutil.ReadFile(path); err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	} else {
		content = string(data)
	}
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

	m := md5.New()
	m.Write(data)
	result := hex.EncodeToString(m.Sum(nil))

	l.Push(lua.LString(result))
	return 1
}

func fileInfo2Table(fi os.FileInfo) *lua.LTable {
	st := fi.Sys().(*syscall.Stat_t)

	var file lua.LTable
	file.RawSetString("filename", lua.LString(fi.Name()))
	file.RawSetString("size", lua.LNumber(fi.Size()))
	file.RawSetString("block_size", lua.LNumber(st.Blksize))
	file.RawSetString("mode", lua.LString(fi.Mode().String()))
	// 001 001 001 = 73 110 100 100 = 420
	file.RawSetString("perm", lua.LNumber(fi.Mode().Perm()))
	file.RawSetString("uid", lua.LNumber(st.Uid))
	file.RawSetString("gid", lua.LNumber(st.Gid))
	file.RawSetString("device", lua.LNumber(st.Dev))
	file.RawSetString("inode", lua.LNumber(st.Ino))
	file.RawSetString("hard_links", lua.LNumber(st.Nlink))
	//file.RawSetString("ctime", lua.LNumber(st.Ctim.Sec))
	//file.RawSetString("mtime", lua.LNumber(st.Mtim.Sec))
	//file.RawSetString("atime", lua.LNumber(st.Atim.Sec))

	typ := "-"
	mod := fi.Mode().String()
	if len(mod) > 9 {
		typ = mod[0:1]
	}
	file.RawSetString("type", lua.LString(typ))

	return &file
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
	lv = l.Get(2)
	if lv != lua.LNil {
		if lv.Type() != lua.LTString {
			l.TypeError(1, lua.LTString)
			return lua.MultRet
		}
		pattern = lv.(lua.LString).String()
	}

	var filearg string
	lv = l.Get(3)
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
			return 2
		}
		errstr := errbuf.String()
		if errstr == "" {
			errstr = err.Error()
		}

		l.Push(lua.LString(""))
		l.Push(lua.LString(errstr))
		return 2
	}
	l.Push(lua.LString(buf.String()))
	l.Push(lua.LString(""))
	return 2
}
