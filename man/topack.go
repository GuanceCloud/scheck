package man

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gobuffalo/packd"
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

// GetAll 返回box中的所有列表 等同于WalkList 会递归返回所有文件列表
func GetAll() []string {
	return ScriptBox.List()
}

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

// WalkList 递归加载文件
func WalkList() {
	ScriptBox.Walk(func(s string, file packd.File) error {

		//fmt.Println(file.Name()) // 二级目录打印结果：users/0500-mysql-weak-psw.lua
		return nil
	})
	//fmt.Println(ScriptBox.List())
}

/*
	清空系统脚本路径
	重新写入脚本

*/
func ScheckCoreSyncDisk(ruleDir string) error {
	//fmt.Println("进入ScheckCoreSyncDisk")
	// 删除目录
	if _, err := os.Stat(ruleDir); err == nil {
		if err := os.RemoveAll(ruleDir); err != nil {
			l.Fatal(err)
			return nil
		}
		//fmt.Println("已经全部删除文件。。")
	}
	// 创建目录，将lua 脚本同步到磁盘上
	if _, err := os.Stat(ruleDir); err != nil {
		if err := os.Mkdir(ruleDir, 0775); err == nil {
			// 遍历 lua脚本名称
			l.Debugf("当前的scriptBox 长度是 %d \n", len(ScriptBox.List()))
			for _, name := range ScriptBox.List() {
				if content, err := ScriptBox.Find(name); err == nil {
					//CreateFile(string(content),fmt.Sprintf("%s/%s"))
					name = strings.ReplaceAll(name, "\\", "/")
					//fmt.Println(strings.Split(name, "/"))
					// 处理多级目录
					paths := strings.Split(name, "/")
					if len(paths) > 1 {
						// 拼接目录
						lib_dir := fmt.Sprintf("%s/%s", ruleDir, strings.Join(paths[0:len(paths)-1], "/"))
						if _, err := os.Stat(lib_dir); err != nil {
							if err := os.MkdirAll(lib_dir, 0775); err != nil {
								l.Fatalf("%s create dir : %s", lib_dir, err)
							}
						}
					}
					// 写文件
					CreateFile(string(content), fmt.Sprintf("%s/%s", ruleDir, name))

				}
			}
		}
	}

	return nil
}

func ScheckDocSyncDisk(path string) error {

	// 创建目录，将lua 脚本同步到磁盘上
	if _, err := os.Stat(path); err != nil {
		if err := os.Mkdir(path, 0775); err == nil {
			// 遍历 lua脚本名称
			l.Debug("当前的scriptBox 长度是 %d \n", len(ScriptBox.List()))
		}
	}
	for _, name := range DocBox.List() {
		if content, err := DocBox.Find(name); err == nil {
			name = strings.ReplaceAll(name, "\\", "/")
			// 处理多级目录
			paths := strings.Split(name, "/")
			if len(paths) > 1 {
				// 拼接目录
				lib_dir := fmt.Sprintf("%s/%s", path, strings.Join(paths[0:len(paths)-1], "/"))
				if _, err := os.Stat(lib_dir); err != nil {
					if err := os.MkdirAll(lib_dir, 0775); err != nil {
						l.Fatalf("%s create dir : %s", lib_dir, err)
					}
				}
			}
			// 写文件
			CreateFile(string(content), fmt.Sprintf("%s/%s", path, name))
		}
	}
	return nil
}

// todo 用户自己的lua文件发生变化时可以自动重载 删除文件后 也要从执行脚本列表中删除?
// ------二次开发使用 从文件夹中读取文件
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
