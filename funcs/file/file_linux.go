//+build linux

package file

import (
	"os"
	"strconv"
	"syscall"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/impl"

	lua "github.com/yuin/gopher-lua"
)

// nolint
func fileInfo2Table(fi os.FileInfo, table *lua.LTable) {
	st := fi.Sys().(*syscall.Stat_t)

	table.RawSetString("filename", lua.LString(fi.Name()))
	table.RawSetString("size", lua.LNumber(fi.Size()))
	table.RawSetString("block_size", lua.LNumber(st.Blksize))
	table.RawSetString("mode", lua.LString(fi.Mode().String()))
	// 001 001 001 = 73 110 100 100 = 420
	table.RawSetString("perm", lua.LNumber(fi.Mode().Perm()))
	table.RawSetString("uid", lua.LNumber(st.Uid))
	table.RawSetString("gid", lua.LNumber(st.Gid))
	table.RawSetString("device", lua.LNumber(st.Dev))
	table.RawSetString("inode", lua.LNumber(st.Ino))
	table.RawSetString("hard_links", lua.LNumber(st.Nlink))
	table.RawSetString("ctime", lua.LNumber(st.Ctim.Sec))
	table.RawSetString("mtime", lua.LNumber(st.Mtim.Sec))
	table.RawSetString("atime", lua.LNumber(st.Atim.Sec))

	if st.Uid == 0 {
		table.RawSetString("username", lua.LString("root"))
	} else {
		if us, err := impl.GetUserDetail(""); err == nil {
			for _, u := range us {
				if u.UID != "0" {
					if uid, err := strconv.Atoi(u.UID); err == nil {
						if uint32(uid) == st.Uid {
							table.RawSetString("username", lua.LString(u.Name))
						}
					}
				}
			}
		}
	}

	typ := "-"
	mod := fi.Mode().String()
	if len(mod) > 9 {
		typ = mod[0:1]
	}
	table.RawSetString("type", lua.LString(typ))
}
