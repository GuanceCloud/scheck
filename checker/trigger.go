package checker

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
)

type provider struct {
	funcs []securityChecker.Func
}

func (p *provider) Funcs() []securityChecker.Func {
	funcs := []securityChecker.Func{
		{Name: `trigger`, Fn: p.trigger},
	}
	return funcs
}

func (p *provider) Catalog() string {
	return "trigger"
}

func (p *provider) trigger(ls *lua.LState) int {

	var err error
	var rule *Rule

	cfg := funcs.GetScriptGlobalConfig(ls)

	if cfg != nil {
		rule = Chk.findRule(cfg.RulePath)
	}

	if rule == nil {
		ls.RaiseError("rule not found")
		return lua.MultRet
	}

	var manifestFileName string
	var templateTable *lua.LTable
	templateVals := map[string]string{}

	lv := ls.Get(1)
	if lv != lua.LNil {
		switch lv.Type() {
		case lua.LTTable:
			templateTable = lv.(*lua.LTable)
		case lua.LTString:
			manifestFileName = string(lv.(lua.LString))
		}
	}

	lv = ls.Get(2) //第一个参数是manifestname时
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
				l.Debugf("type %s ignored: %s", v.Type().String(), v.String())
			}
		})
	}

	if manifestFileName == "" {
		//use the default manifest
		manifestFileName = strings.TrimSuffix(filepath.Base(rule.File), filepath.Ext(rule.File))
	}
	manifest, err := Chk.getManifest(manifestFileName)
	if err != nil {
		ls.RaiseError("%s", err)
		return lua.MultRet
	}

	fields := map[string]interface{}{}
	tags := map[string]string{}
	tm := time.Now().UTC()

	tags["title"] = manifest.Title
	tags["level"] = manifest.Level
	tags["category"] = manifest.Category
	tags["version"] = git.Version
	for k, v := range manifest.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	message := manifest.Desc

	if manifest.tmpl != nil {
		buf := bytes.NewBufferString("")
		if err = manifest.tmpl.Execute(buf, templateVals); err != nil {
			ls.RaiseError("fail to apple template, %s", err)
			return lua.MultRet
		} else {
			message = buf.String()
		}
	}
	fields["message"] = message

	if err = output.SendMetric(manifest.RuleID, tags, fields, tm); err != nil {
		ls.RaiseError("%s", err)
		return lua.MultRet
	}

	return 0
}
func firstTrigger() {
	tags := map[string]string{}
	tm := time.Now().UTC()

	tags["title"] = "scheck start"
	tags["level"] = "-"
	tags["category"] = "system"
	tags["version"] = git.Version
	fields := map[string]interface{}{}
	cronNum, intervalNum := Chk.scheduler.countInfo()
	luas := cronNum + intervalNum
	formatTime := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("scheck 程序启动在%s上 \n 当前共%d个lua进入巡检队列 时间：%s", luas, formatTime)
	fields["message"] = message
	_ = output.SendMetric("0000-scheck-start", tags, fields, tm)

}
func init() {
	securityChecker.AddFuncProvider(&provider{})
}
