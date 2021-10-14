package lua

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/config"

	"github.com/influxdata/toml"
	"github.com/influxdata/toml/ast"
	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"
)

func TestLua(rulepath string) {
	rulepath, _ = filepath.Abs(rulepath)
	if !strings.HasSuffix(rulepath, global.LuaExt) {
		rulepath += global.LuaExt
	}

	byteCode, err := CompilesScript(rulepath)
	if err != nil {
		fmt.Printf("Compile lua scripterr=%v \n", err)
		return
	}
	if config.Cfg != nil {
		lua.LuaPathDefault = filepath.Join(config.Cfg.System.RuleDir, "lib", "?.lua")
		lua.LuaPathDefault += ";" + filepath.Join(config.Cfg.System.CustomRuleLibDir, "?.lua")
	} else {
		lua.LuaPathDefault = filepath.Join(global.InstallDir, global.DefRulesDir, "lib", "?.lua")
	}

	ls := lua.NewState(lua.Options{SkipOpenLibs: true})
	if err = LoadLuaLibs(ls); err != nil {
		ls.Close()
		fmt.Printf("LoadLuaLibs err=%v \n ", err)
		return
	}
	utils.SetScriptGlobalConfig(ls, &utils.ScriptGlobalCfg{RulePath: rulepath})
	LoadModule(ls)
	lfunc := ls.NewFunctionFromProto(byteCode.Proto)
	ls.Push(lfunc)
	if err = ls.PCall(0, lua.MultRet, nil); err != nil {
		fmt.Printf("testLua err=%v \n", err)
	}
}

// CheckLua check all custom lua.
func CheckLua(customRuleDir string) {
	fileInfos, err := ioutil.ReadDir(customRuleDir)
	if err != nil {
		l.Errorf("%v", err)
		return
	}
	if len(fileInfos) == 0 {
		fmt.Printf("there are no lua rules here %s \n", customRuleDir)
		return
	}
	errCount := 0
	for _, info := range fileInfos {
		if info.IsDir() {
			continue
		}
		if strings.HasSuffix(info.Name(), ".lua") {
			_, err := CompilesScript(filepath.Join(customRuleDir, info.Name()))
			if err != nil {
				fmt.Printf("name of lua :%s compiles is err:%v \n", info.Name(), err)
				errCount++
			}
		}
		if strings.HasSuffix(info.Name(), ".manifest") {
			err := CompilesManifest(filepath.Join(customRuleDir, info.Name()))
			if err != nil {
				fmt.Printf("name of manifest :%s compiles is err:%v \n", info.Name(), err)
				errCount++
			}
		}
	}
	if errCount != 0 {
		fmt.Printf("there are %d error here \n", errCount)
	} else {
		fmt.Printf("all of the lua rules is ok! \n")
	}
}

func CompilesManifest(fileName string) error {
	var tbl *ast.Table
	contents, err := ioutil.ReadFile(fileName)
	if err != nil {
		return err
	}
	tbl, err = toml.Parse(contents)
	if err != nil {
		return err
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
	for k := range requireKeys {
		v := tbl.Fields[k]
		if v == nil {
			continue
		}
		str := ""
		if kv, ok := v.(*ast.KeyValue); ok {
			if s, ok := kv.Value.(*ast.String); ok {
				str = s.Value
			}
			if s, ok := kv.Value.(*ast.Array); ok {
				str = s.Source()
			}
		}
		if str != "" {
			requireKeys[k] = true
		}
	}
	for name, ok := range requireKeys {
		if !ok {
			return fmt.Errorf("field name=%s can find", name)
		}
	}
	return nil
}
