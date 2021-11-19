//go:build windows
// +build windows

package file

import (
	"os"

	lua "github.com/yuin/gopher-lua"
)

func fileInfo2Table(fi os.FileInfo, table *lua.LTable) {
	typ := "-"
	mod := fi.Mode().String()
	if len(mod) > 9 {
		typ = mod[0:1]
	}
	table.RawSetString("type", lua.LString(typ))
}
