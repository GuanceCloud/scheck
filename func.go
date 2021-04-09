package securityChecker

import (
	"fmt"
	"io"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type FuncImp func(*lua.LState) int

type Func struct {
	Name string
	Fn   FuncImp
}

type FuncProvider interface {
	Catalog() string
	Funcs() []Func
}

var FuncProviders []FuncProvider

func AddFuncProvider(p FuncProvider) {
	FuncProviders = append(FuncProviders, p)
}

func DumpSupportLuaFuncs(w io.Writer) {
	names := []string{}
	for _, p := range FuncProviders {
		for _, f := range p.Funcs() {
			names = append(names, fmt.Sprintf("%s", f.Name))
		}
	}
	s := strings.Join(names, "\n")
	s += "\n"
	w.Write([]byte(s))
}
