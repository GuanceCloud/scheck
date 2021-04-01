package checker

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"

	"github.com/influxdata/toml"
)

// Rule corresponding to a lua script file
type Rule struct {
	File  string
	Proto *lua.FunctionProto

	Hash       string
	LastModify time.Duration

	Manifests []*RuleManifest

	cron string //多个manifest的cron应当一样
}

type RuleManifest struct {
	RuleID   string `toml:"id"`
	Category string `toml:"category"`
	Level    string `toml:"level"`
	Title    string `toml:"title"`
	Desc     string `toml:"desc"`
	Cron     string `toml:"cron"`

	Tags map[string]string `toml:"tags,omitempty"`

	filename string

	tmpl *template.Template
}

func (c *Rule) toLuaTable() *lua.LTable {
	var t lua.LTable
	t.RawSetString("rulefile", lua.LString(c.File))
	return &t
}

func (r *Rule) run(ls *lua.LState) error {
	defer func() {
		if e := recover(); e != nil {
			log.Errorf("panic, %v", e)
		}
	}()
	var err error
	lfunc := ls.NewFunctionFromProto(r.Proto)
	ls.Push(lfunc)
	err = ls.PCall(0, lua.MultRet, nil)
	ls.Pop(ls.GetTop())
	return err
}

func (c *Checker) loadManifest(name string, r *Rule) (*RuleManifest, error) {
	if !strings.HasSuffix(name, ".manifest") {
		name += ".manifest"
	}

	for _, r := range r.Manifests {
		if r.filename == name {
			return r, nil
		}
	}

	data, err := ioutil.ReadFile(filepath.Join(c.rulesDir, name))
	if err != nil {
		log.Errorf("%s", err)
		return nil, err
	}

	var rm RuleManifest
	if err := toml.Unmarshal(data, &rm); err != nil {
		log.Errorf("%s", err)
		return nil, err
	}

	if rm.Desc == "" {
		err = fmt.Errorf("desc of manifst cannot be empty")
		log.Errorf("%s", err)
		return nil, err
	}

	if rm.tmpl, err = template.New("test").Parse(rm.Desc); err != nil {
		log.Errorf("invalid desc: %s", err)
		return nil, err
	}

	if rm.RuleID == "" {
		err = fmt.Errorf("id of manifst cannot be empty")
		log.Errorf("%s", err)
		return nil, err
	}

	rm.filename = name

	r.Manifests = append(r.Manifests, &rm)

	return &rm, nil
}

func (c *Checker) newRuleFromFile(luapath string) (*Rule, error) {

	luaname := strings.TrimRight(filepath.Base(luapath), filepath.Ext(luapath))

	proto, err := compileLua(luapath)
	if err != nil {
		return nil, err
	}

	r := &Rule{
		File:  luapath,
		Proto: proto,
	}

	rm, err := c.loadManifest(luaname, r)
	if err != nil {
		return nil, err
	}

	if _, err := specParser.Parse(rm.Cron); err != nil {
		log.Errorf("invalid cron: %s, %s", rm.Cron, err)
		return nil, err
	}

	r.cron = rm.Cron

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
		log.Errorf("fail to parse lua file '%s', err: %s", filePath, err)
		return nil, err
	}
	proto, err := lua.Compile(chunk, filePath)
	if err != nil {
		log.Errorf("fail to compile lua file '%s', err: %s", filePath, err)
		return nil, err
	}
	return proto, nil
}
