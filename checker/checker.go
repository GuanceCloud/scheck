package checker

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	luajson "github.com/layeh/gopher-json"
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
		rules    []*Rule
		lStates  []*luaState

		cron *luaCron

		luaExtends *luaext.LuaExt
	}

	luaState struct {
		lState *lua.LState
		rule   *Rule
	}
)

func (c *Checker) findRule(rulefile string) *Rule {
	if rulefile == "" {
		return nil
	}
	for _, r := range c.rules {
		if r.File == rulefile {
			return r
		}
	}
	return nil
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

	if err := c.loadFiles(); err != nil {
		return
	}

	if len(c.rules) == 0 {
		log.Warnf("no rule found")
	}

	c.cron.start()

	for _, r := range c.rules {
		ls := c.newLuaState(r)
		if ls != nil {
			c.lStates = append(c.lStates, ls)
			c.cron.addLuaScript(ls)
		}
	}

	<-ctx.Done()
	c.cron.stop()
}

// Start
func Start(ctx context.Context, rulesDir, outputpath string) {

	checker = &Checker{
		rulesDir:   rulesDir,
		luaExtends: luaext.NewLuaExt(),
		cron:       newLuaCron(),
	}

	log.Debugf("output: %s", outputpath)

	output.NewOutputer(outputpath)
	checker.start(ctx)
}

func (c *Checker) loadFiles() error {

	ls, err := ioutil.ReadDir(c.rulesDir)
	if err != nil {
		log.Errorf("%s", err)
		return err
	}

	for _, f := range ls {
		if f.IsDir() {
			continue
		}

		path := filepath.Join(c.rulesDir, f.Name())

		if !strings.HasSuffix(f.Name(), ".lua") {
			continue
		}

		if r, err := c.newRuleFromFile(path); err == nil {
			c.rules = append(c.rules, r)
		}
	}

	return nil
}

func (c *Checker) newLuaState(r *Rule) *luaState {
	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	ls.SetGlobal(luaext.OneFileGlobalConfKey, r.toLuaTable())
	if err := c.registerLua(ls); err != nil {
		return nil
	}
	return &luaState{
		lState: ls,
		rule:   r,
	}
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

	var manifest *RuleManifest
	if manifestFileName == "" {
		if len(rule.Manifests) == 0 {
			l.Push(lua.LString("manifest not found"))
			return 1
		}
		manifest = rule.Manifests[0]
	} else {
		if manifest, err = c.loadManifest(manifestFileName, rule); err != nil {
			l.Push(lua.LString(fmt.Sprintf("manifest of '%s' not found", manifestFileName)))
			return 1
		}
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
