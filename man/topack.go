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
	Others      = map[string]interface{}{
		"apis": "man/manuals/apis.md", //...
	}
)

// GetAllName 返回配置中的所有lua列表 去除后缀名的格式
func GetAllName() []string {
	all := ScriptBox.List()
	rms := make([]string, 0)
	for _, name := range all {
		//index := strings.LastIndex(name, ".")
		index := path.Ext(name)
		if index != ".manifest" { // 没有后缀名
			continue
		}
		rms = append(rms, strings.TrimSuffix(name, ".manifest"))
	}

	return rms
}

// todo 用户自己的lua文件发生变化时可以自动重载 删除文件后 也要从执行脚本列表中删除?
// ------二次开发使用 从文件夹中读取文件
func GetTpl(box *packr.Box, name string) (string, error) {
	return box.FindString(name)
}

func GetManifest(name string) (string, error) {
	return ScriptBox.FindString(name + ".manifest")
}
