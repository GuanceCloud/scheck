package funcs

import (
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
