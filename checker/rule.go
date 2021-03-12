package checker

import (
	"bufio"
	"os"
	"time"

	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
)

type Rule struct {
	File  string
	Proto *lua.FunctionProto

	Interval time.Duration
	LastRun  time.Time
}

func (r *Rule) run(ls *lua.LState) error {
	var err error
	lfunc := ls.NewFunctionFromProto(r.Proto)
	ls.Push(lfunc)
	err = ls.PCall(0, lua.MultRet, nil)
	ls.Pop(ls.GetTop())
	return err
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
		l.Errorf("fail to parse file: %s,  err:%s", filePath, err)
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		l.Errorf("fail to compile file: %s,  err:%s", filePath, err)
		return nil, err
	}
	return proto, nil
}

func newRuleFromFile(path string) (*Rule, error) {

	proto, err := compileLua(path)
	if err != nil {
		return nil, err
	}

	r := &Rule{
		File:  path,
		Proto: proto,
	}

	if r.Interval == 0 {
		r.Interval = time.Second * 10
	}
	return r, nil
}
