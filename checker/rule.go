package checker

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

// Rule corresponding to a lua script file
type Rule struct {
	File  string
	Proto *lua.FunctionProto

	LastModify int64

	Manifests []*RuleManifest

	cron string //多个manifest的cron应当一样

	mux sync.Mutex

	disabled bool

	markAsDelete int32
	stopch       chan bool
	running      int32

	lState *lua.LState
	cronID int
}

type RuleManifest struct {
	RuleID   string `toml:"id"`
	Category string `toml:"category"`
	Level    string `toml:"level"`
	Title    string `toml:"title"`
	Desc     string `toml:"desc"`
	Cron     string `toml:"cron"`

	Tags map[string]string

	path string

	tmpl *template.Template

	disabled bool

	LastModify int64
}

func (r *Rule) reload() {

	r.mux.Lock()
	defer r.mux.Unlock()

	fi, err := os.Stat(r.File)
	if err != nil {
		log.Errorf("%s", err)
		return
	} else {
		if fi.ModTime().Unix() > int64(r.LastModify) {
			log.Debugf("%s changed, reload it", r.File)

			proto, err := compileLua(r.File)
			if err != nil {
				return
			}
			r.Proto = proto
			r.LastModify = fi.ModTime().Unix()
		}
	}

	for _, m := range r.Manifests {

		fi, err = os.Stat(m.path)
		if err != nil {
			log.Errorf("%s", err)
			continue
		} else {
			if fi.ModTime().Unix() <= m.LastModify {
				continue
			}
		}

		log.Debugf("%s changed, reload it", m.path)

		if nm, err := loadManifest(m.path); err != nil {
			log.Errorf("%s", err)
			continue
		} else {
			*m = *nm
			r.disabled = nm.disabled
			r.cron = nm.Cron
		}
		m.LastModify = fi.ModTime().Unix()
	}
}

func (r *Rule) findManifest(path string) *RuleManifest {
	r.mux.Lock()
	defer r.mux.Unlock()

	if path == "" {
		if len(r.Manifests) > 0 {
			return r.Manifests[0]
		}
	} else {
		for _, m := range r.Manifests {
			if m.path == path {
				return m
			}
		}
		if nm, err := loadManifest(path); err != nil {
			log.Errorf("%s", err)
			return nil
		} else {
			r.Manifests = append(r.Manifests, nm)
			return nm
		}
	}
	return nil
}

func (r *Rule) toLuaTable() *lua.LTable {
	var t lua.LTable
	t.RawSetString("rulefile", lua.LString(r.File))
	return &t
}

func (r *Rule) run() error {
	if r.lState == nil {
		return nil
	}
	var err error
	r.mux.Lock()
	lfunc := r.lState.NewFunctionFromProto(r.Proto)
	r.mux.Unlock()
	r.lState.Push(lfunc)
	err = r.lState.PCall(0, lua.MultRet, nil)
	r.lState.Pop(r.lState.GetTop())
	return err
}

func loadManifest(path string) (*RuleManifest, error) {

	var rm RuleManifest
	rm.path = path
	err := rm.parse()
	if err != nil {
		log.Errorf("%s", err)
		return nil, err
	}
	rm.LastModify = time.Now().Unix()

	return &rm, nil
}

func ensureFieldString(k string, v interface{}, s *string) error {
	if kv, ok := v.(*ast.KeyValue); ok {
		if str, ok := kv.Value.(*ast.String); ok {
			*s = str.Value
			return nil
		}
	}
	return fmt.Errorf("unknown value for field '%s', expecting string", k)
}

func ensureFieldBool(k string, v interface{}, b *bool) error {
	var err error
	if kv, ok := v.(*ast.KeyValue); ok {
		switch t := kv.Value.(type) {
		case *ast.Boolean:
			*b, err = t.Boolean()
			if err != nil {
				return fmt.Errorf("unknown boolean value type %q, expecting boolean", kv.Value)
			}
			return nil
		case *ast.String:
			*b, err = strconv.ParseBool(t.Value)
			if err != nil {
				return fmt.Errorf("unknown boolean value type %q, expecting boolean", kv.Value)
			}
			return nil
		}
	}
	return fmt.Errorf("unknown value for field '%s', expecting boolean", k)
}

func (rm *RuleManifest) parse() error {

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("parse panic, %v", e)
		}
	}()

	contents, err := ioutil.ReadFile(rm.path)
	if err != nil {
		return err
	}

	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))
	tbl, err := toml.Parse(contents)
	if err != nil {
		return err
	}

	mustKeys := map[string]bool{
		"id":       false,
		"category": false,
		"level":    false,
		"title":    false,
		"desc":     false,
		"cron":     false,
	}

	for k := range mustKeys {
		v := tbl.Fields[k]
		if v == nil {
			continue
		}
		str := ""
		if err := ensureFieldString(k, v, &str); err != nil {
			return err
		} else {
			switch k {
			case "id":
				rm.RuleID = str
			case "category":
				rm.Category = str
			case "level":
				rm.Level = str
			case "title":
				rm.Title = str
			case "desc":
				rm.Desc = str
			case "cron":
				rm.Cron = str
			}
			if str != "" {
				mustKeys[k] = true
			}
		}
	}

	for k, bset := range mustKeys {
		if !bset {
			return fmt.Errorf("%s must not be empty", k)
		}
	}

	if rm.tmpl, err = template.New("test").Parse(rm.Desc); err != nil {
		return fmt.Errorf("invalid desc: %s", err)
	}

	if _, err := specParser.Parse(rm.Cron); err != nil {
		return fmt.Errorf("invalid cron: %s, %s", rm.Cron, err)
	}

	rm.Tags = map[string]string{}
	for k, v := range tbl.Fields {
		if _, ok := mustKeys[k]; ok {
			continue
		}

		if v == nil {
			continue
		}

		if k == "disabled" {
			bval := false
			if err := ensureFieldBool(k, v, &bval); err != nil {
				return err
			}
			rm.disabled = bval
			continue
		}

		str := ""
		err = ensureFieldString(k, v, &str)
		if err != nil {
			return err
		}
		if str != "" {
			rm.Tags[k] = str
		}
	}
	return nil
}

func (c *Checker) newRuleFromFile(rulename, dir string) (*Rule, error) {

	luapath := filepath.Join(dir, rulename+".lua")
	manifestPath := filepath.Join(dir, rulename+".manifest")

	proto, err := compileLua(luapath)
	if err != nil {
		return nil, err
	}

	r := &Rule{
		File:   luapath,
		Proto:  proto,
		stopch: make(chan bool),
	}

	rm, err := loadManifest(manifestPath)
	if err != nil {
		return nil, err
	}
	r.Manifests = append(r.Manifests, rm)

	fi, err := os.Stat(luapath)
	if err == nil {
		r.LastModify = fi.ModTime().Unix()
	}
	fi, err = os.Stat(manifestPath)
	if err == nil {
		rm.LastModify = fi.ModTime().Unix()
	}

	r.cron = rm.Cron
	r.disabled = rm.disabled

	return r, nil
}

func compileLua(filePath string) (*lua.FunctionProto, error) {
	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		log.Errorf("fail to open %s, error:%s", filePath, err)
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
