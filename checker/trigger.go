package checker

import (
	"bytes"
	"fmt"
	"os"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/luafuncs"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/output"
)

func Trigger(ls *lua.LState) int {
	cfg := utils.GetScriptGlobalConfig(ls)
	var manifestFileName string
	var templateTable *lua.LTable
	templateVals := map[string]string{}
	lv := ls.Get(global.LuaArgIdx1)
	if lv != lua.LNil {
		switch lv.Type() { // nolint
		case lua.LTTable:
			var ok bool
			templateTable, ok = lv.(*lua.LTable)
			if !ok {
				l.Warnf("type assertion checked error")
			}
		case lua.LTString:
			manifestFileName = string(lv.(lua.LString))
		default:
		}
	}

	lv = ls.Get(global.LuaArgIdx2)
	if lv.Type() == lua.LTTable {
		var ok bool
		templateTable, ok = lv.(*lua.LTable)
		if !ok {
			l.Warnf("type assertion checked error")
		}
	}

	if templateTable != nil {
		templateTable.ForEach(func(k lua.LValue, v lua.LValue) {
			switch v.Type() { // nolint
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
		manifestFileName = cfg.RulePath
	}

	manifest, err := GetManifestByName(manifestFileName)
	if err != nil {
		ls.RaiseError("%s", err)
		return lua.MultRet
	}
	fields := map[string]interface{}{}
	tm := time.Now().UTC()
	message := manifest.Desc
	tags := makeManifestTags(manifest)
	if manifest.tmpl != nil {
		buf := bytes.NewBufferString("")
		if err = manifest.tmpl.Execute(buf, templateVals); err != nil {
			ls.RaiseError("fail to apple template, %s", err)
			return lua.MultRet
		}
		message = buf.String()
	}
	fields["message"] = message
	if err = output.SendMetric(manifest.RuleID, tags, fields, tm); err != nil {
		ls.RaiseError("%s", err)
		return lua.MultRet
	}
	go luafuncs.UpdateTriggerCount(cfg.RulePath)
	return 0
}

func makeManifestTags(manifest *RuleManifest) map[string]string {
	tags := map[string]string{}
	tags["title"] = manifest.Title
	tags["level"] = manifest.Level
	tags["category"] = manifest.Category
	tags["version"] = git.Version
	for k, v := range manifest.tags {
		if _, ok := tags[k]; !ok {
			tags[k] = v
		}
	}
	return tags
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
	message := fmt.Sprintf("scheck started, %d rules ready at %s", luas, formatTime)
	fields["message"] = message
	_ = output.SendMetric("0000-scheck-start", tags, fields, tm)
}
