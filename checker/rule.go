package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
)

// Rule corresponding to a lua script file
type Rule struct {
	File string

	byteCode   *funcs.ByteCode
	lastModify int64

	cron string //多个manifest的cron应当一样

	mux sync.Mutex

	disabled bool

	markAsDelete int32
	stopch       chan bool
	running      int32

	rt     *funcs.ScriptRunTime
	cronID int
}

type RuleManifest struct {
	RuleID   string `toml:"id"`
	Category string `toml:"category"`
	Level    string `toml:"level"`
	Title    string `toml:"title"`
	Desc     string `toml:"desc"`
	Cron     string `toml:"cron"`

	tags map[string]string

	path string

	tmpl *template.Template

	disabled bool

	lastModify int64
}

func newRule(path string) *Rule {
	return &Rule{
		File: path,
	}
}

func (r *Rule) load() error {

	r.mux.Lock()
	defer r.mux.Unlock()

	fi, err := os.Stat(r.File)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	if fi.ModTime().Unix() > r.lastModify {
		if r.lastModify > 0 {
			log.Debugf("%s changed, reload it", r.File)
		}

		bcode, err := funcs.CompilesScript(r.File)
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		r.byteCode = bcode
		r.lastModify = fi.ModTime().Unix()
	}

	//load default manifest for cron info
	ruledir := filepath.Dir(r.File)
	rulename := strings.TrimSuffix(filepath.Base(r.File), filepath.Ext(r.File))
	manifestPath := filepath.Join(ruledir, rulename+".manifest")

	manifest := checker.manifests[manifestPath]
	if manifest == nil {
		manifest = newManifest(manifestPath)
		checker.addManifest(manifest)
	}

	if err = manifest.load(); err != nil {
		err = fmt.Errorf("fail to load %s, %s", manifestPath, err)
		log.Errorf("%s", err)
		return err
	}
	r.cron = manifest.Cron
	r.disabled = manifest.disabled

	return nil
}

func (r *Rule) run() {
	if atomic.LoadInt32(&r.running) > 0 {
		return
	}
	if r.disabled {
		return
	}

	defer func() {
		if e := recover(); e != nil {
			log.Errorf("panic, %v", e)
		}
		atomic.AddInt32(&r.running, -1)
		close(r.stopch)
	}()
	atomic.AddInt32(&r.running, 1)
	r.stopch = make(chan bool)

	var err error
	if r.rt == nil {
		if r.rt, err = funcs.GetScriptRuntime(&funcs.ScriptGlobalCfg{
			RulePath: r.File,
		}); err != nil {
			log.Errorf("%s", err)
			return
		}
	}

	log.Debugf("start run %s", filepath.Base(r.File))
	r.mux.Lock()
	err = r.rt.PCall(r.byteCode)
	r.mux.Unlock()
	if err != nil {
		log.Errorf("run %s failed, %s", filepath.Base(r.File), err)
	} else {
		log.Debugf("run %s ok", filepath.Base(r.File))
	}
}

func newManifest(path string) *RuleManifest {
	return &RuleManifest{
		path: path,
	}
}

func (m *RuleManifest) load() error {

	fi, err := os.Stat(m.path)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	if fi.ModTime().Unix() > m.lastModify {
		if m.lastModify > 0 {
			log.Debugf("%s changed, reload it", m.path)
		} else {
			log.Debugf("load manifest: %s", m.path)
		}
		err := m.parse()
		if err != nil {
			log.Errorf("%s", err)
			return err
		}
		m.lastModify = time.Now().Unix()
	}

	return nil
}

func (m *RuleManifest) parse() (err error) {

	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("parse panic, %v", e)
			log.Errorf("%s", err)
		}
	}()

	rm := *m

	var contents []byte
	var tbl *ast.Table
	contents, err = ioutil.ReadFile(rm.path)
	if err != nil {
		return
	}
	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))
	tbl, err = toml.Parse(contents)
	if err != nil {
		return
	}

	requireKeys := map[string]bool{
		"id":       false,
		"category": false,
		"level":    false,
		"title":    false,
		"desc":     false,
		"cron":     false,
	}

	for k := range requireKeys {
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
				if str == "" {
					str = Cfg.Cron
				}
				rm.Cron = str
			}
			if str != "" {
				requireKeys[k] = true
			}
		}
	}

	for k, bset := range requireKeys {
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

	rm.tags = map[string]string{}
	omithost := false
	hostname := ""
	for k, v := range tbl.Fields {
		if _, ok := requireKeys[k]; ok {
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
		} else if k == "omit_hostname" {
			bval := false
			if err := ensureFieldBool(k, v, &bval); err != nil {
				return err
			}
			omithost = bval
			continue
		} else if k == "hostname" {
			str := ""
			err = ensureFieldString(k, v, &str)
			if err != nil {
				return err
			}
			hostname = str
		}

		str := ""
		err = ensureFieldString(k, v, &str)
		if err != nil {
			return err
		}
		if str != "" {
			rm.tags[k] = str
		}
	}

	if !omithost {
		if hostname == "" {
			if h, err := os.Hostname(); err == nil {
				hostname = h
			}
		}
		rm.tags["host"] = hostname
	}

	*m = rm
	return nil
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
