//+build windows

package system

import (
	"os"

	lua "github.com/yuin/gopher-lua"
)

func fileInfo2Table(fi os.FileInfo) *lua.LTable {
	var file lua.LTable
	var fileModeL = 9
	//file.RawSetString("filename", lua.LString(fi.Name()))
	//file.RawSetString("size", lua.LNumber(fi.Size()))
	//file.RawSetString("block_size", lua.LNumber(st.Blksize))
	//file.RawSetString("mode", lua.LString(fi.Mode().String()))
	//// 001 001 001 = 73 110 100 100 = 420
	//file.RawSetString("perm", lua.LNumber(fi.Mode().Perm()))
	//file.RawSetString("uid", lua.LNumber(st.Uid))
	//file.RawSetString("gid", lua.LNumber(st.Gid))
	//file.RawSetString("device", lua.LNumber(st.Dev))
	//file.RawSetString("inode", lua.LNumber(st.Ino))
	//file.RawSetString("hard_links", lua.LNumber(st.Nlink))
	//file.RawSetString("ctime", lua.LNumber(st.Ctim.Sec))
	//file.RawSetString("mtime", lua.LNumber(st.Mtim.Sec))
	//file.RawSetString("atime", lua.LNumber(st.Atim.Sec))

	typ := "-"
	mod := fi.Mode().String()
	if len(mod) > fileModeL {
		typ = mod[0:1]
	}
	file.RawSetString("type", lua.LString(typ))

	return &file
}
