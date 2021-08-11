package man

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr/v2"
)

var (
	ScriptBox   = packr.New("libs", "./libs")
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

// todo 检查 、加载、预编译 都在这里做
func checkFile(extName string) (luaf, mf string) {
	luaf, _ = GetLua(extName)
	mf, _ = GetManifest(extName)
	return
}

// WalkList 递归加载文件
func WalkList() {
	ScriptBox.Walk(func(s string, file packd.File) error {

		fmt.Println(file.Name()) // 二级目录打印结果：users/0500-mysql-weak-psw.lua
		return nil
	})
}

/*
	清空系统脚本路径
	重新写入脚本
	-- todo 删除rule.d目录 重新写入时 判断os arch
*/
func ScheckCoreSyncDisk(ruleDir string) error {
	log.Println("进入ScheckCoreSyncDisk")
	// 删除目录
	if _, err := os.Stat(ruleDir); err == nil {
		if err := os.RemoveAll(ruleDir); err != nil {
			log.Fatal(err)
			return nil
		}
		fmt.Println("已经全部删除文件。。")
	}
	// 创建目录，将lua 脚本同步到磁盘上
	if _, err := os.Stat(ruleDir); err != nil {
		if err := os.Mkdir(ruleDir, 0775); err == nil {
			// 遍历 lua脚本名称
			log.Printf("当前的scriptBox 长度是 %d \n", len(ScriptBox.List()))
			exts := make(map[string]string)
			for _, name := range ScriptBox.List() {
				// list是 lib/123.lua  001-add.lua
				if content, err := ScriptBox.Find(name); err == nil {
					name = strings.ReplaceAll(name, "\\", "/")
					paths := strings.Split(name, "/")
					if len(paths) > 1 {
						// 拼接目录
						lib_dir := fmt.Sprintf("%s/%s", ruleDir, strings.Join(paths[0:len(paths)-1], "/"))
						if _, err := os.Stat(lib_dir); err != nil {
							if err := os.MkdirAll(lib_dir, 0775); err != nil {
								log.Fatalf("%s create dir : %s", lib_dir, err)
							}
						}
						CreateFile(string(content), fmt.Sprintf("%s/%s", ruleDir, name))
					} else { //目录是一级
						extName := path.Ext(name)
						if _, ok := exts[extName]; !ok {
							luaF, mf := checkFile(extName)
							if luaF != "" && mf != "" {
								CreateFile(luaF, fmt.Sprintf("%s/%s.%s", ruleDir, extName, "lua"))
								CreateFile(mf, fmt.Sprintf("%s/%s.%s", ruleDir, extName, "manifest"))
							}
						} else {
							continue
						}
						// 写文件

					}

				}
			}
			fmt.Println(err)
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
