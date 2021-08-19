package man

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var (
	ScriptBox   = packr.New("libs", "./libs")
	DocBox      = packr.New("doc", "./doc")
	TemplateBox = packr.New("template", "./template")
)

// GetAllName
func GetAllName() []string {
	all := ScriptBox.List()
	rms := make([]string, 0)
	for _, name := range all {
		index := path.Ext(name)
		if index != ".manifest" { // 没有后缀名
			continue
		}
		rms = append(rms, strings.TrimSuffix(name, ".manifest"))
	}

	return rms
}

/*
	remove all rule files
	Write the rule in scriptBox to the file
*/
func ScheckCoreSyncDisk(ruleDir string) {
	if _, err := os.Stat(ruleDir); err == nil {
		if err := os.RemoveAll(ruleDir); err != nil {
			l.Fatal(err)
		}
	}

	if _, err := os.Stat(ruleDir); err != nil {
		if err := os.Mkdir(ruleDir, 0775); err == nil {
			for _, name := range ScriptBox.List() {
				if content, err := ScriptBox.Find(name); err == nil {
					name = strings.ReplaceAll(name, "\\", "/")
					paths := strings.Split(name, "/")
					if len(paths) > 1 {
						lib_dir := fmt.Sprintf("%s/%s", ruleDir, strings.Join(paths[0:len(paths)-1], "/"))
						if _, err := os.Stat(lib_dir); err != nil {
							if err := os.MkdirAll(lib_dir, 0775); err != nil {
								l.Fatalf("%s create dir : %s", lib_dir, err)
							}
						}
					}
					CreateFile(string(content), fmt.Sprintf("%s/%s", ruleDir, name))

				}
			}
		}
	}

}

func ScheckDocSyncDisk(path string) error {

	if _, err := os.Stat(path); err != nil {
		if err := os.Mkdir(path, 0775); err == nil {
			// 遍历 lua脚本名称
			l.Debug("the scriptBox lens= %d \n", len(ScriptBox.List()))
		}
	}
	for _, name := range DocBox.List() {
		if content, err := DocBox.Find(name); err == nil {
			name = strings.ReplaceAll(name, "\\", "/")
			paths := strings.Split(name, "/")
			if len(paths) > 1 {
				lib_dir := fmt.Sprintf("%s/%s", path, strings.Join(paths[0:len(paths)-1], "/"))
				if _, err := os.Stat(lib_dir); err != nil {
					if err := os.MkdirAll(lib_dir, 0775); err != nil {
						l.Fatalf("%s create dir : %s", lib_dir, err)
					}
				}
			}
			CreateFile(string(content), fmt.Sprintf("%s/%s", path, name))
		}
	}
	return nil
}

func GetMD(name string) (string, error) {
	return ScriptBox.FindString(name + ".md")
}

func GetLua(name string) (string, error) {
	return ScriptBox.FindString(name + ".lua")
}

func GetTpl(name string) (string, error) {
	return TemplateBox.FindString(name + ".tpl")
}

func GetManifest(name string) (string, error) {
	return ScriptBox.FindString(name + ".manifest")
}
