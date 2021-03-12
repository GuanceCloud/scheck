package funcs

import lua "github.com/yuin/gopher-lua"

type (
	LuaFunc struct {
		Name string
		Desc string
		Fn   lua.LGFunction
	}
)

var SupportFuncs = []LuaFunc{
	{
		Name: `exist`,
		Fn:   fileExist,
		Desc: `
boolean exist(filepath)
check file exist.`,
	},

	{
		Name: `file`,
		Fn:   readFile,
		Desc: `
string file(filepath)
reads and return file content, it issues an error when read failed.`,
	},
}
