package checker

import (
	"bytes"
	"fmt"
	"os"
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
	return []securityChecker.Func{
		{Name: `trigger`, Fn: p.trigger},
	}
}

func (p *provider) Catalog() string {
	return "trigger"
}

func (p *provider) trigger(ls *lua.LState) int {

	cfg := funcs.GetScriptGlobalConfig(ls)

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

	lv = ls.Get(2)
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
		manifestFileName = cfg.RulePath
	}

	manifest, err := GetManifestByName(manifestFileName)
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
	tags["level"] = "info"
	tags["category"] = "system"
	tags["version"] = git.Version
	if h, err := os.Hostname(); err == nil {
		tags["host"] = h
	}
	fields := map[string]interface{}{}

	luas := GetRuleNum()
	formatTime := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("scheck 程序启动，当前共%d个lua进入巡检队列，启动时间为：%s", luas, formatTime)
	fields["message"] = message
	_ = output.SendMetric("0000-scheck-start", tags, fields, tm)

}
func init() {
	securityChecker.AddFuncProvider(&provider{})
}
