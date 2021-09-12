package checker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

	lua "github.com/yuin/gopher-lua"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"
)

// Rule corresponding to a lua script file
type Rule struct {
	File     string
	Name     string
	byteCode *luafuncs.ByteCode
	cron     string
	mux      sync.Mutex
	disabled bool
	interval int64 // 间隔时间
	RunTime  int64 // 下一次执行时间 单位秒
	manifest *RuleManifest
}

type RuleManifest struct {
	RuleID     string   `toml:"id"`
	Category   string   `toml:"category"`
	Level      string   `toml:"level"`
	Title      string   `toml:"title"`
	Desc       string   `toml:"desc"`
	Cron       string   `toml:"cron"`
	OSArch     []string `toml:"os_arch"`
	tags       map[string]string
	path       string
	tmpl       *template.Template
	disabled   bool
	lastModify int64
}

func newRule(path string) *Rule {
	return &Rule{
		File: path,
	}
}

// load 从文件夹中加载
func (r *Rule) load() error {
	r.mux.Lock()
	defer r.mux.Unlock()

	bcode, err := luafuncs.CompilesScript(r.File)
	if err != nil {
		l.Errorf("%s", err)
		return err
	}
	r.byteCode = bcode

	// load default manifest for cron info
	ruledir := filepath.Dir(r.File)
	rulename := strings.TrimSuffix(filepath.Base(r.File), filepath.Ext(r.File))
	r.Name = rulename
	manifestPath := filepath.Join(ruledir, rulename+".manifest")

	manifest := r.manifest
	if manifest == nil {
		manifest = newManifest(manifestPath)
		r.manifest = manifest
	}

	if err = manifest.load(); err != nil {
		l.Errorf("fail to load %s, %s", manifestPath, err)
		return err
	}

	// 添加操作系统参数字段后 需要判断是否运行该lua文件
	runLua := false
	for _, localOS := range manifest.OSArch {
		if strings.Contains(strings.ToUpper(localOS), strings.ToUpper(runtime.GOOS)) {
			runLua = true
		}
	}
	if !runLua {
		return fmt.Errorf(" The OS:%s cannot load this manifest :%s ", runtime.GOOS, r.Name)
	}

	r.cron = manifest.Cron
	if r.cron == "" || r.cron == global.LuaCronDisable {
		r.interval = -1
	} else {
		r.interval = checkRunTime(r.cron)
	}
	r.disabled = manifest.disabled
	r.RunTime = time.Now().UnixNano()/1e6 + r.interval
	return nil
}

func (r *Rule) RunJob(state *luafuncs.ScriptRunTime) {
	now := time.Now()

	// to set filePath
	var lt lua.LTable
	lt.RawSetString(global.LuaConfigurationKey, lua.LString(r.Name))
	state.Ls.SetGlobal(global.LuaConfiguration, &lt)

	cxt, cancel := context.WithTimeout(context.Background(), global.LuaScriptTimeout)
	defer cancel()
	state.Ls.SetContext(cxt)

	lFunc := state.Ls.NewFunctionFromProto(r.byteCode.Proto)
	state.Ls.Push(lFunc)
	errChan := make(chan bool)
	var err error
	go func() {
		l.Debugf("rule name: %s is running!!!", r.Name)
		if err = state.Ls.PCall(0, lua.MultRet, nil); err != nil {
			errChan <- false
		} else {
			errChan <- true
		}
	}()
	select {
	case <-cxt.Done():
		l.Errorf("run lua script:%s is timeout!", r.Name)
	case b := <-errChan:
		if !b {
			l.Errorf("lua.state run  err=%v ", err)
		}
	}
	go luafuncs.UpdateStatus(r.Name, time.Since(now), err != nil)

	state.Ls.RemoveContext()

	pool.putPool(state)
}

func (r *Rule) RunOnce(cxt context.Context, c chan string) {
	if pool == nil {
		l.Warn("the statePool is nil!!!")
		return
	}
	state := getNewState()
	state.Ls.SetContext(cxt)

	var lt lua.LTable
	lt.RawSetString(global.LuaConfigurationKey, lua.LString(r.Name))
	state.Ls.SetGlobal(global.LuaConfiguration, &lt)

	l.Debugf("rule name: %s is running at once", r.Name)
	lFunc := state.Ls.NewFunctionFromProto(r.byteCode.Proto)
	state.Ls.Push(lFunc)
	if err := state.Ls.PCall(0, lua.MultRet, nil); err != nil {
		l.Errorf("lua.state run  err=%v ", err)
		c <- r.Name
	}
}

func newManifest(path string) *RuleManifest {
	return &RuleManifest{
		path: path,
	}
}

func (rm *RuleManifest) load() error {
	fi, err := os.Stat(rm.path)
	if err != nil {
		l.Errorf("%s", err)
		return err
	}

	if fi.ModTime().Unix() > rm.lastModify {
		if rm.lastModify > 0 {
			l.Debugf("%s changed, reload it", rm.path)
		} else {
			l.Debugf("load manifest: %s", rm.path)
		}
		err := rm.parse()
		if err != nil {
			l.Errorf("%s", err)
			return err
		}
		rm.lastModify = time.Now().Unix()
	}
	return nil
}

func (rm *RuleManifest) parse() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("parse panic, %v", e)
			l.Errorf("%s", err)
		}
	}()
	rm1 := *rm
	var contents []byte
	var tbl *ast.Table
	contents, err = ioutil.ReadFile(rm1.path)
	if err != nil {
		l.Warnf("read file err=%v", err)
		return
	}
	// 去掉有可能在UTF8编码中存在的BOM头
	contents = bytes.TrimPrefix(contents, []byte("\xef\xbb\xbf"))
	tbl, err = toml.Parse(contents)
	if err != nil {
		l.Warnf("toml.Parse err=%v", err)
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
	// 屏蔽字段
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

	rm1.setRequireKey(tbl, requireKeys)

	for k, bset := range requireKeys {
		if !bset {
			return fmt.Errorf("%s must not be empty", k)
		}
	}
	// 模版rm.Desc
	if rm1.tmpl, err = template.New("test").Parse(rm1.Desc); err != nil {
		return fmt.Errorf("invalid desc: %s", err)
	}

	if _, err = specParser.Parse(rm1.Cron); err != nil {
		return fmt.Errorf("invalid cron: %s, %s", rm1.Cron, err)
	}

	if err := rm1.setTag(tbl, requireKeys, invalidField); err != nil {
		return err
	}

	*rm = rm1
	return nil
}

func (rm *RuleManifest) setRequireKey(tbl *ast.Table, requireKeys map[string]bool) {
	for k := range requireKeys {
		v := tbl.Fields[k]
		if v == nil {
			continue
		}
		str := ""
		err := ensureFieldString(k, v, &str)
		if err != nil {
			return
		}
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
				str = config.Cfg.System.Cron
			}
			rm.Cron = str
		case "os_arch":
			arr, err := ensureFieldStrings(k, v)
			if err != nil {
				l.Warnf("os_arch is err = %v", err)
			}
			rm.OSArch = arr
		}
		if str != "" {
			requireKeys[k] = true
		}
	}
}

func (rm *RuleManifest) setTag(tbl *ast.Table, requireKeys, invalidField map[string]bool) error {
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
			err := ensureFieldString(k, v, &str)
			if err != nil {
				return err
			}
			hostname = str
		}

		str := ""
		err := ensureFieldString(k, v, &str)
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
			hostname, _ = os.Hostname()
		}
		rm.tags["host"] = hostname
	}
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

func ensureFieldStrings(k string, v interface{}) ([]string, error) {
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
func checkInterval(cronStr string) int64 {
	fields := strings.Fields(cronStr)
	intervals := map[int]int64{}

	for idx, f := range fields {
		parts := strings.Split(f, "/")
		if len(parts) == 2 && parts[0] == "*" {
			interval, err := strconv.ParseInt(parts[1], global.ParseBase, global.ParseBitSize)
			if err == nil && interval > 0 {
				intervals[idx] = interval
			}
		} else if f != "*" {
			return 0
		}
	}
	timeDurations := []int64{
		0: int64(time.Second),
		1: int64(time.Minute),
		2: int64(time.Hour),
		3: int64(time.Hour) * 24,
		4: int64(time.Hour) * 24 * 30,
	}
	if len(intervals) == 1 {
		for k, v := range intervals {
			return v * timeDurations[k]
		}
	}
	return 0
}

var cronMaps = map[int]int64{
	0: 1, // second
	1: 60,
	2: 60 * 60,
	3: 60 * 60 * 24,
	4: 60 * 60 * 24 * 30,
}
var cronInterval = []int64{60, 60, 24, 30, 1, 1}

func checkRunTime(cronStr string) int64 {
	nextRunTime := int64(0)
	fields := strings.Fields(cronStr)
	for idx, f := range fields {
		parts := strings.Split(f, "/")
		if len(parts) == 2 && parts[0] == "*" {
			interval, err := strconv.ParseInt(parts[1], global.ParseBase, global.ParseBitSize)
			if err == nil && interval > 0 {
				nextRunTime += (cronMaps[idx]) * interval
			}
		}
	}
	// 1 1 2 * *
	if nextRunTime == 0 {
		swap := int64(0)
		for idx, f := range fields {
			if f != "*" {
				swap = cronMaps[idx] * cronInterval[idx]
			}
		}
		nextRunTime = swap
	}
	return nextRunTime * global.TimeMilli
}
