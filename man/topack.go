package man

import (
	"path"
	"strings"

	"github.com/gobuffalo/packr/v2"
)

var (
	ScriptBox   = packr.New("libs", "./libs")
	DocBox      = packr.New("doc", "./doc")
	TemplateBox = packr.New("template", "./template")
)

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

func GetTpl(box *packr.Box, name string) (string, error) {
	return box.FindString(name)
}

func GetManifest(name string) (string, error) {
	return ScriptBox.FindString(name + ".manifest")
}
