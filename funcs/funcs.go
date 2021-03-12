package funcs

import (
	"bytes"
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type (
	LuaFunc struct {
		Name  string
		Fn    lua.LGFunction
		Title string
		Desc  string
		Test  []string
	}
)

var SupportFuncs = []LuaFunc{
	{
		Name:  `exist`,
		Fn:    fileExist,
		Title: `boolean exist(filepath)`,
		Desc: `
check file exist.`,
		Test: []string{`
file='demo.txt'
if exist(file) then
	print(file.." exists")
else
	print(file.." not exists")
end
`},
	},

	{
		Name:  `file`,
		Fn:    readFile,
		Title: `string file(filepath)`,
		Desc: `
reads and return file content, it issues an error when read failed.`,
		Test: []string{},
	},
}

func DumpSupports(showDesc, showDemo bool) string {

	s := bytes.NewBufferString("")

	for _, f := range SupportFuncs {
		s.WriteString(fmt.Sprintf("%s", strings.TrimSpace(f.Title)))
		s.WriteString("\n")
		if showDesc {
			s.WriteString(fmt.Sprintf("  %s", strings.TrimSpace(f.Desc)))
			s.WriteString("\n\n")
		}

		if showDemo {
			for idx, t := range f.Test {
				s.WriteString(fmt.Sprintf("  Demo %d:", idx+1))
				s.WriteString(t)
			}
			s.WriteString("\n\n")
		}

	}
	fmt.Printf("%s\n", s.String())
	return s.String()
}
