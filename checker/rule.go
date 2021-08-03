package checker

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"text/template"
	"time"

	log "github.com/sirupsen/logrus"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
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

	rt         *funcs.ScriptRunTime
	scheduleID int

	interval time.Duration
}

type RuleManifest struct {
	RuleID   string   `toml:"id"`
	Category string   `toml:"category"`
	Level    string   `toml:"level"`
	Title    string   `toml:"title"`
	Desc     string   `toml:"desc"`
	Cron     string   `toml:"cron"`
	OSArch   []string `toml:"os_arch"`
	tags     map[string]string

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

	manifest := Chk.manifests[manifestPath]
	if manifest == nil {
		manifest = newManifest(manifestPath)
		Chk.addManifest(manifest)
	}

	if err = manifest.load(); err != nil {
		err = fmt.Errorf("fail to load %s, %s", manifestPath, err)
		log.Errorf("%s", err)
		return err
	}
	reschedule := false
	if r.cron != "" && manifest.Cron != r.cron {
		log.Debugf("cron change from %s to %s", r.cron, manifest.Cron)
		Chk.scheduler.DelRule(r) //
		reschedule = true
	}
	// 添加操作系统参数字段后 需要判断是否运行该lua文件
	runLua := false
	for _, localOS := range manifest.OSArch {
		if strings.Contains(strings.ToUpper(localOS), strings.ToUpper(runtime.GOOS)) {
			//log.Warnf("有相同的操作系统匹配到。。。 localos=%s", localOS)
			runLua = true
		}
	}
	if !runLua {
		//log.Warnf("manifest中不支持当前操作系统 当前的os=%s", strings.ToUpper(runtime.GOOS))
		Chk.doDelRule(r) // 删除当前规则 删除lua文件
		return fmt.Errorf("manifest中不支持当前操作系统")
	}

	r.cron = manifest.Cron
	r.interval = checkInterval(r.cron)
	r.disabled = manifest.disabled
	if reschedule {
		Chk.reSchedule(r)
	}

	return nil
}

func (r *Rule) run() {
	if r == nil {
		return
	}
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
		log.Warnf("read file err=%v", err)
		return
	}
	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))
	tbl, err = toml.Parse(contents)
	if err != nil {
		log.Warnf("toml.Parse err=%v", err)
		return
	}

	requireKeys := map[string]bool{
		"id":       false,
		"category": false,
		"level":    false,
		"title":    false,
		"desc":     false,
		"cron":     false,
		"os_arch":  false,
	}
	//屏蔽字段
	invalidField := map[string]bool{
		"description":  false,
		"riskitems":    false,
		"audit":        false,
		"remediation":  false,
		"impact":       false,
		"defaultvalue": false,
		"rationale":    false,
		"references":   false,
		"CIS":          false,
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
					str = config.Cfg.Cron
				}
				rm.Cron = str
			case "os_arch":
				fmt.Printf("-------str=%s \n", str)
				arr, err := ensureFieldStrings(k, v, &str)
				if err != nil {
					log.Warnf("获取os_arch字段失败err = %v", err)
				}
				rm.OSArch = arr

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

	// 模版rm.Desc
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
			_, ok := invalidField[k]
			if !ok {
				rm.tags[k] = str
			}
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
		if str, ok := kv.Value.(*ast.Array); ok {

			*s = str.Source()
			return nil
		}
	}

	return fmt.Errorf("unknown value for field '%s', expecting string", k)
}
func ensureFieldStrings(k string, v interface{}, s *string) ([]string, error) {
	arr := make([]string, 0)
	if kv, ok := v.(*ast.KeyValue); ok {
		if str, ok := kv.Value.(*ast.Array); ok {
			for _, val := range str.Value {
				arr = append(arr, val.Source())
			}
			return arr, nil
		}
	}

	return nil, fmt.Errorf("unknown value for field '%s', expecting string", k)
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

// 从cron中 取出设置的间隔时间
func checkInterval(cronStr string) time.Duration {
	fields := strings.Fields(cronStr)
	intervals := map[int]int64{}

	for idx, f := range fields {
		parts := strings.Split(f, "/")
		if len(parts) == 2 && parts[0] == "*" {
			interval, err := strconv.ParseInt(parts[1], 10, 64)
			if err == nil && interval > 0 {
				intervals[idx] = interval
			}
		} else {
			if f != "*" {
				return 0
			}
		}
	}

	if len(intervals) == 1 {
		for k, v := range intervals {
			switch k {
			case 0:
				return time.Duration(v * int64(time.Second))
			case 1:
				return time.Duration(v * int64(time.Minute))
			case 2:
				return time.Duration(v * int64(time.Hour))
			case 3:
				return time.Duration(v * int64(time.Hour) * 24)
			case 4:
				return time.Duration(v * int64(time.Hour) * 24 * 30)
			}
		}
	}

	return 0
}
