package checker

import (
	"bytes"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	lua "github.com/yuin/gopher-lua"

	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"
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

func (p *provider) trigger(l *lua.LState) int {

	var err error
	var rule *Rule

	cfg := funcs.GetScriptGlobalConfig(l)
	if cfg != nil {
		rule = Chk.findRule(cfg.RulePath)
	}

	if rule == nil {
		l.RaiseError("rule not found")
		return lua.MultRet
	}

	var manifestFileName string
	var templateTable *lua.LTable
	templateVals := map[string]string{}

	lv := l.Get(1)
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

	if manifestFileName == "" {
		//use the default manifest
		manifestFileName = strings.TrimSuffix(filepath.Base(rule.File), filepath.Ext(rule.File))
	}
	manifest, err := Chk.getManifest(manifestFileName)
	if err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	fields := map[string]interface{}{}
	tags := map[string]string{}
	tm := time.Now().UTC()

	tags["title"] = manifest.Title
	tags["level"] = manifest.Level
	tags["category"] = manifest.Category
	for k, v := range manifest.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	message := manifest.Desc

	if manifest.tmpl != nil {
		buf := bytes.NewBufferString("")
		if err = manifest.tmpl.Execute(buf, templateVals); err != nil {
			l.RaiseError("fail to apple template, %s", err)
			return lua.MultRet
		} else {
			message = buf.String()
		}
	}
	fields["message"] = message

	if err = output.SendMetric(manifest.RuleID, tags, fields, tm); err != nil {
		l.RaiseError("%s", err)
		return lua.MultRet
	}

	return 0
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
