package checker

import (
	"bufio"
	"log"
	"os"
	"time"

	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
)

// Rule corresponding to a lua script file
type Rule struct {
	File  string
	Proto *lua.FunctionProto

	Hahs       string
	LastModify time.Duration

	Interval time.Duration
	LastRun  time.Time
}

// Run a rule in lua virtual machine
func (r *Rule) Run(ls *lua.LState) error {
	defer func() {
		if e := recover(); e != nil {
			log.Printf("[panic] %v", e)
		}
	}()
	var err error
	lfunc := ls.NewFunctionFromProto(r.Proto)
	ls.Push(lfunc)
	err = ls.PCall(0, lua.MultRet, nil)
	ls.Pop(ls.GetTop())
	return err
}

//NewRuleFromFile construct a Rule struct from lua file
func NewRuleFromFile(path string) (*Rule, error) {

	proto, err := compileLua(path)
	if err != nil {
		return nil, err
	}

	r := &Rule{
		File:     path,
		Proto:    proto,
		Interval: 30 * time.Second,
	}

	return r, nil
}

func compileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(file)
	chunk, err := luaparse.Parse(reader, filePath)
	if err != nil {
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		return nil, err
	}
	return proto, nil
}
