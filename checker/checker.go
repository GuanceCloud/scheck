package checker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	luajson "github.com/layeh/gopher-json"
	cron "github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/luaext"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
)

var (
	checker *Checker
)

type (
	Checker struct {
		rulesDir string
		rules    map[string]*Rule
		//lStates  []*luaState

		cron *luaCron

		luaExtends *luaext.LuaExt

		mux sync.Mutex

		loading int32
	}
)

func newChecker(rulesDir string) *Checker {
	c := &Checker{
		rules:      map[string]*Rule{},
		rulesDir:   rulesDir,
		luaExtends: luaext.NewLuaExt(),
		cron:       newLuaCron(),
	}
	return c
}

func (c *Checker) findRule(rulefile string) *Rule {
	if rulefile == "" {
		return nil
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if r, ok := c.rules[rulefile]; ok {
		return r
	}
	return nil
}
func (c *Checker) addRule(r *Rule) {
	c.mux.Lock()
	defer c.mux.Unlock()

	if err := c.getLuaState(r); err == nil {
		if id, err := c.cron.addLuaScript(r); err != nil {
			log.Errorf("%s", err)
			r.lState.Close()
		} else {
			r.cronID = id
			c.rules[r.File] = r
		}
	} else {
		log.Errorf("%s", err)
	}
}

func (c *Checker) delRules() {
	c.mux.Lock()
	defer c.mux.Unlock()

	for _, r := range c.rules {
		if atomic.LoadInt32(&r.markAsDelete) == 0 {
			continue
		}
		log.Debugf("removing %s", r.File)
		if atomic.LoadInt32(&r.running) > 0 {
			select {
			case <-r.stopch:
				c.doDelRule(r)
			case <-time.After(time.Second * 10):
				log.Warnf("remove rule failed, timeout")
			}
		} else {
			c.doDelRule(r)
		}

	}
}

func (c *Checker) doDelRule(r *Rule) {
	c.cron.Cron.Remove(cron.EntryID(r.cronID))
	if r.lState != nil {
		if !r.lState.IsClosed() {
			r.lState.Close()
		}
	}
	delete(c.rules, r.File)
}

func (c *Checker) start(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			log.Errorf("panic %s", e)
			log.Errorf("%s", string(buf[:n]))

		}
		output.Outputer.Close()
		log.Info("checker exit")
	}()

	log.Debugf("rule dir: %s", c.rulesDir)

	c.cron.start()

	if err := c.loadRules(ctx, c.rulesDir); err != nil {
		return
	}

	select {
	case <-ctx.Done():
		return
	default:
	}

	if len(c.rules) == 0 {
		log.Warnf("no rule loaded")
	}

	go func() {
		for {
			sleepContext(ctx, time.Second*10)
			log.Debugf("reload rules")

			select {
			case <-ctx.Done():
				return
			default:
			}

			c.loadRules(ctx, c.rulesDir)
		}
	}()

	<-ctx.Done()
	c.cron.stop()
}

// Start
func Start(ctx context.Context, rulesDir, outputpath string) {

	checker = newChecker(rulesDir)

	log.Debugf("output: %s", outputpath)

	output.NewOutputer(outputpath)
	checker.start(ctx)
}

func (c *Checker) loadRules(ctx context.Context, ruleDir string) error {

	if atomic.LoadInt32(&c.loading) > 0 {
		return nil
	}
	atomic.AddInt32(&c.loading, 1)
	defer atomic.AddInt32(&c.loading, -1)
	ls, err := ioutil.ReadDir(ruleDir)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	for _, r := range c.rules {
		atomic.AddInt32(&r.markAsDelete, 1)
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		path := filepath.Join(ruleDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			continue
		}

		if exist, ok := c.rules[path]; ok {
			exist.reload()
			atomic.AddInt32(&exist.markAsDelete, -1)
			continue
		}

		rulename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))

		if r, err := c.newRuleFromFile(rulename, ruleDir); err == nil {
			c.addRule(r)
		}
	}

	c.delRules()

	return nil
}

func (c *Checker) getLuaState(r *Rule) error {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	ls.SetGlobal(luaext.OneFileGlobalConfKey, r.toLuaTable())
	if err := c.registerLua(ls); err != nil {
		return err
	}
	r.lState = ls
	return nil
}

func (c *Checker) registerLua(lstate *lua.LState) error {
	luaext.LuaExtendFuncs = append(luaext.LuaExtendFuncs,
		luaext.LuaFunc{
			Name: `trig`,
			Fn:   c.trig,
		})
	if err := luaext.LoadLuaLibs(lstate); err != nil {
		log.Errorf("%s", err)
		return err
	}
	luajson.Preload(lstate) //for json parse
	for _, f := range luaext.LuaExtendFuncs {
		lstate.Register(f.Name, f.Fn)
	}
	return nil
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	if duration == 0 {
		return nil
	}

	t := time.NewTimer(duration)
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		t.Stop()
		return ctx.Err()
	}
}

func (c *Checker) trig(l *lua.LState) int {

	var err error
	var rule *Rule

	lv := l.GetGlobal(luaext.OneFileGlobalConfKey)
	if lv.Type() == lua.LTTable {
		t := lv.(*lua.LTable)
		rulefile := t.RawGetString("rulefile").String()
		rule = c.findRule(rulefile)
	}

	if rule == nil {
		l.Push(lua.LString("rule not found"))
		return 1
	}

	manifestFileName := ""

	var templateTable *lua.LTable
	templateVals := map[string]string{}

	lv = l.Get(1)
	if lv != lua.LNil {
		switch lv.Type() {
		case lua.LTTable:
			templateTable = lv.(*lua.LTable)
		case lua.LTString:
			manifestFileName = string(lv.(lua.LString))
		}
	}

	lv = l.Get(2) //第一个参数是manifestname时
	if lv.Type() == lua.LTTable {
		templateTable = lv.(*lua.LTable)
	}

	if templateTable != nil {
		templateTable.ForEach(func(k lua.LValue, v lua.LValue) {
			switch v.Type() {
			case lua.LTBool:
			case lua.LTNumber:
			case lua.LTString:
				templateVals[k.String()] = v.String()
			default:
				log.Debugf("type %s ignored: %s", v.Type().String(), v.String())
			}
		})
	}

	var mpath string
	if manifestFileName != "" {
		if !strings.HasPrefix(manifestFileName, ".manifest") {
			manifestFileName += ".manifest"
		}
		mpath = filepath.Join(c.rulesDir, manifestFileName)
	}
	manifest := rule.findManifest(mpath)
	if manifest == nil {
		l.Push(lua.LString(fmt.Sprintf("manifest of '%s' not found", manifestFileName)))
		return 1
	}

	fields := map[string]interface{}{}
	tags := map[string]string{}
	tm := time.Now().UTC()

	tags["title"] = manifest.Title
	tags["level"] = manifest.Level
	tags["category"] = manifest.Category
	for k, v := range manifest.Tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	message := manifest.Desc

	if manifest.tmpl != nil {
		buf := bytes.NewBufferString("")
		if err = manifest.tmpl.Execute(buf, templateVals); err != nil {
			l.Push(lua.LString(fmt.Sprintf("fail to apple template, %s", err)))
			return 1
		} else {
			message = buf.String()
		}
	}
	fields["message"] = message

	if err = output.Outputer.SendMetric(manifest.RuleID, tags, fields, tm); err != nil {
		l.Push(lua.LString(fmt.Sprintf("%s", err)))
		return 1
	}

	l.Push(lua.LString(""))
	return 1
}

func TestRule(rulepath string) {

	rulename := filepath.Base(rulepath)
	ruledir := filepath.Dir(rulepath)

	c := newChecker(ruledir)
	output.NewOutputer("")

	r, err := c.newRuleFromFile(rulename, ruledir)
	if err != nil {
		return
	}
	c.rules[r.File] = r
	if err := c.getLuaState(r); err != nil {
		return
	} else {
		if err := r.run(); err != nil {
			log.Errorf("%s", err)
		}
	}

	return
}
