package utils

import (
	"bufio"

	//nolint:gosec
	"crypto/md5"

	//nolint:gosec
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	lua "github.com/yuin/gopher-lua"
)

func HashMd5(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	str := string(lv.(lua.LString))
	f, err := os.Open(str) //nolint:gosec
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	defer f.Close() // nolint:errcheck,gosec
	r := bufio.NewReader(f)

	h := md5.New() //nolint:gosec

	_, err = io.Copy(h, r) //nolint
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	sum := h.Sum(nil)
	l.Push(lua.LString(fmt.Sprintf("%X", sum)))
	return 1
}

func HashSha256(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	str := string(lv.(lua.LString))
	f, err := os.Open(str) //nolint:gosec
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	defer f.Close() // nolint:errcheck
	r := bufio.NewReader(f)

	ha := sha256.New()

	_, err = io.Copy(ha, r) //nolint
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	sum := ha.Sum(nil)
	l.Push(lua.LString(fmt.Sprintf("%X", sum)))
	return 1
}

func HashSha1(l *lua.LState) int {
	lv := l.Get(1)
	if lv.Type() != lua.LTString {
		l.TypeError(1, lua.LTString)
		return lua.MultRet
	}
	str := string(lv.(lua.LString))
	f, err := os.Open(str) //nolint:gosec
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}
	defer f.Close() // nolint:errcheck
	r := bufio.NewReader(f)

	ha := sha1.New() //nolint:gosec

	_, err = io.Copy(ha, r) //nolint
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	sum := ha.Sum(nil)
	l.Push(lua.LString(fmt.Sprintf("%X", sum)))
	return 1
}
