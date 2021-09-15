//+build windows

package system

import (
	"os"

	lua "github.com/yuin/gopher-lua"
)

func fileInfo2Table(fi os.FileInfo) *lua.LTable {
	var file lua.LTable
	var fileModeL = 9
	typ := "-"
	mod := fi.Mode().String()
	if len(mod) > fileModeL {
		typ = mod[0:1]
	}
	file.RawSetString("type", lua.LString(typ))

	return &file
}
