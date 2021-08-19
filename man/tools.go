package man

/*
模版模块
从manifest中读取文件 后生成md格式的模版 导出命令：-doc 使用
*/
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"

	"github.com/BurntSushi/toml"
)

var (
	l = logger.SLogger("tool")
)

func SetLog() {
	l = logger.DefaultSLogger("tool")
}

type Tmp struct {
	Path string
	Obj  interface{}
}

/*
list to string
*/
func (t *Tmp) arrayTostring(i []string) string {
	var str = ""
	for k, v := range i {
		if k == 0 {
			str += v
		} else {
			str += fmt.Sprintf(",%s", v)
		}
	}
	return str
}

func (t *Tmp) htmlUrl(i []string) string {
	str := ""
	for _, v := range i {
		title := ""
		url := ""
		spaceRe, _ := regexp.Compile("\\[(.*)\\]")
		// matched, err := regexp.MatchString("[^\\[\\]\\(\\)]+", s)
		matches := spaceRe.FindAllStringSubmatch(v, -1)
		if matches != nil {
			for _, v := range matches {
				title = string(v[0][1 : len(v[0])-1])
			}
			spaceRe, _ := regexp.Compile("(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]")
			// matched, err := regexp.MatchString("[^\\[\\]\\(\\)]+", s)
			matches := spaceRe.FindAllStringSubmatch(v, -1)
			if matches != nil {
				url = matches[0][0]
				str += fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>\n<br/>", url, title)
			}
		} else {
			spaceRe, _ := regexp.Compile("(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]")
			matches := spaceRe.FindAllStringSubmatch(v, -1)
			if matches != nil {
				url = matches[0][0]
				str += fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>\n", url, url)
			} else {
				str += "无"
			}
		}
	}
	return str
}

func (t *Tmp) GetTemplate() string {
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"inc":     t.arrayTostring,
		"htmlUrl": t.htmlUrl,
	}

	tpl, err := GetTpl(t.Path)
	tmpl, err := template.New(t.Path).Funcs(funcMap).Parse(tpl)
	if err != nil {
		l.Fatalf("parsing: %s", err)
	}
	buf := new(bytes.Buffer)

	err = tmpl.Execute(buf, t.Obj)
	if err != nil {
		l.Fatalf("execution: %s", err)
	}
	return buf.String()
}

type Markdown struct {
	RuleID       string   `toml:"id"`
	Category     string   `toml:"category"`
	Level        string   `toml:"level"`
	Title        string   `toml:"title"`
	Desc         string   `toml:"desc"`
	Cron         string   `toml:"cron"`
	Disabled     bool     `toml:"disabled"`
	OSArch       []string `toml:"os_arch"`
	Description  []string `toml:"description"`
	Rationale    []string `toml:"rationale"`
	Riskitems    []string `toml:"riskitems"`
	Audit        string   `toml:"audit"`
	Remediation  string   `toml:"remediation"`
	Impact       []string `toml:"impact"`
	Defaultvalue []string `toml:"defaultvalue"`
	References   []string `toml:"references"`
	Cis          []string `toml:"CIS"`
	Id           string
	path         string
}

func (m *Markdown) TemplateDecodeFile() error {
	fileStr, err := GetManifest(m.path)
	if err != nil {
		l.Warnf("没有此manifest文件 name=%s", m.path)
		return err
	}
	if err = toml.Unmarshal([]byte(fileStr), m); err != nil {
		l.Warnf("反序列化错误 err=%v", err)
		return err
	}
	return nil
}

func PathExists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ScheckDir(id string, outputPath string) {
	path := fmt.Sprintf("%s/%s", outputPath, id)

	bool, _ := PathExists(path)
	if !bool {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			l.Fatalf("%s 创建失败", path)
		}
	}
}

func IsAppand(file_path string) bool {
	files, _ := ioutil.ReadFile(file_path)
	if strings.Index(string(files), "fitOs") >= 0 {
		return true
	}
	return false
}

func ScheckList(dir_path string) []string {
	files, _ := ioutil.ReadDir(dir_path)
	manifest := make([]string, 0)
	for _, f := range files {
		scId := strings.Split(f.Name(), ".manifest")
		if path.Ext(f.Name()) == ".manifest" {
			if IsAppand(fmt.Sprintf("%s%s", dir_path, f.Name())) {
				manifest = append(manifest, scId[0])
			}
		}
	}
	return manifest
}

func CreateFile(content string, file string) error {
	file = doFilepath(file)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		l.Fatalf("open file err=%v", err)
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		l.Fatalf("write to file err=%v", err)
		return err
	}
	return nil
}
func doFilepath(file string) string {
	return strings.ReplaceAll(file, "\\", "/")

}

type Summary struct {
	Category map[string]map[string]string
}

func DfTemplate(filesName []string, outputPath string) {
	if _, err := os.Stat(outputPath); err != nil {
		if err := os.MkdirAll(outputPath, 0775); err != nil {
			l.Fatalf("%s create dir : %s", outputPath, err)
		}
	}
	count := 0
	for _, v := range filesName {
		yamlPath := fmt.Sprintf("%s/%s/manifest.yaml", outputPath, v)
		metaPath := fmt.Sprintf("%s/%s/meta.md", outputPath, v)
		id := strings.Split(v, "-")[0]
		md := Markdown{path: v, Id: id}
		md.TemplateDecodeFile()
		if len(md.Description) < 1 {
			continue
		}
		yaml := Tmp{Path: "manifest", Obj: md}
		meta := Tmp{Path: "md", Obj: md}
		ScheckDir(v, outputPath)
		CreateFile(yaml.GetTemplate(), yamlPath)
		CreateFile(meta.GetTemplate(), metaPath)
		count++
	}
	l.Debugf("模版生成  manifest数量=%d , 在%s 目录下", count, outputPath)

}

func ToMakeMdFile(filesName []string, outputPath string) {
	if _, err := os.Stat(outputPath); err != nil {
		if err := os.MkdirAll(outputPath, 0775); err != nil {
			l.Fatalf("%s create dir : %s", outputPath, err)
		}
	}
	category := map[string]map[string]string{
		"system":    make(map[string]string),
		"network":   make(map[string]string),
		"storage":   make(map[string]string),
		"container": make(map[string]string),
		"db":        make(map[string]string),
	}

	count := 0
	for _, v := range filesName {
		id := strings.Split(v, "-")[0]
		md := Markdown{path: v, Id: id}
		if err := md.TemplateDecodeFile(); err != nil {
			l.Fatalf("err:%s", err)
		}

		if len(md.Description) < 1 {
			continue
		}
		_, ok := category[md.Category]
		if ok {
			category[md.Category][fmt.Sprintf("%s-%s", strings.Split(md.RuleID, "-")[0], md.Title)] = md.RuleID
		}
		yuquemd := Tmp{Path: "yuquemd", Obj: md}
		if _, err := os.Stat(outputPath); err != nil {
			if err := os.MkdirAll(outputPath, 0775); err != nil {
				l.Fatalf("%s create dir : %s", outputPath, err)
			}
		}
		yuqueMdPath := fmt.Sprintf("%s/%s.md", outputPath, md.RuleID)
		err := CreateFile(yuquemd.GetTemplate(), yuqueMdPath)
		if err != nil {
			l.Fatalf("createFile err=%v \n", err)
		}
		count++
	}

	yuque := Tmp{Path: "summary", Obj: Summary{Category: category}}
	yuquePath := fmt.Sprintf("%s/%s", outputPath, "summary.md")
	l.Debugf("doc文档生成  manifest数量=%d , 在%s 目录下 \n", count, outputPath)
	ScheckDocSyncDisk(outputPath)
	CreateFile(yuque.GetTemplate(), yuquePath)

}
