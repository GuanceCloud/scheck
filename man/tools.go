package man

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/internal/global"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/git"

	"github.com/BurntSushi/toml"
	"github.com/gobuffalo/packr/v2"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

var (
	l          = logger.DefaultSLogger("tool")
	regexpStr  = "\\[(.*)\\]"
	regexpHTTP = "(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]"
)

type Params struct {
	Version        string
	ReleaseDate    string
	AvailableArchs string
}

type Tmp struct {
	Path string
	Obj  interface{}
	box  *packr.Box
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

func (t *Tmp) htmlURL(i []string) string {
	str := ""
	for _, v := range i {
		title := ""
		url := ""
		spaceRe := regexp.MustCompile(regexpStr)

		matches := spaceRe.FindAllStringSubmatch(v, -1)
		if matches != nil {
			for _, v := range matches {
				title = v[0][1 : len(v[0])-1]
			}
			spaceRe, _ := regexp.Compile(regexpHTTP)
			matches := spaceRe.FindAllStringSubmatch(v, -1)
			if matches != nil {
				url = matches[0][0]
				str += fmt.Sprintf("<a href=\"%s\" target=\"_blank\">%s</a>\n<br/>", url, title)
			}
		} else {
			spaceRe, _ := regexp.Compile(regexpHTTP)
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
		"htmlUrl": t.htmlURL,
	}
	// Create template, add template function, add parsing template text
	tpl, err := GetTpl(t.box, t.Path)
	if err != nil {
		l.Fatalf("GetTpl err=%s", err.Error())
	}
	newTmpl, err := template.New(t.Path).Funcs(funcMap).Parse(tpl)
	if err != nil {
		l.Fatalf("parsing: %s", err.Error())
	}
	buf := new(bytes.Buffer)

	err = newTmpl.Execute(buf, t.Obj)
	if err != nil {
		l.Fatalf("execution: %s", err.Error())
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
	ID           string
	path         string
}

func (m *Markdown) TemplateDecodeFile() error {
	fileStr, err := GetManifest(m.path)
	if err != nil {
		l.Warnf("There is no such manifest file name=%s", m.path)
		return err
	}
	if err = toml.Unmarshal([]byte(fileStr), m); err != nil {
		l.Warnf("Deserialization error err=%v", err)
		return err
	}
	return nil
}

/*
	Determine whether the file or folder exists
	If the error returned is nil, the file or folder exists
	If the returned error type is judged as true using OS. Isnotexist(), the file or folder does not exist
	If the error returned is of another type, it is uncertain whether it exists
*/
func PathExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ScheckDir(id, outputPath string) {
	filePath := fmt.Sprintf("%s/%s", outputPath, id)

	exist, _ := PathExists(filePath)
	if !exist {
		err := os.Mkdir(filePath, os.ModePerm)
		if err != nil {
			l.Fatalf("%s Creation failed", filePath)
		}
	}
}

func IsAppand(filePath string) bool {
	files, _ := ioutil.ReadFile(filePath)
	return bytes.Contains(files, []byte("fitOs"))
}

func ScheckList(dirPath string) []string {
	files, _ := ioutil.ReadDir(dirPath)
	manifest := make([]string, 0)
	for _, f := range files {
		scID := strings.Split(f.Name(), global.LuaManifestExt)
		if path.Ext(f.Name()) == global.LuaManifestExt {
			if IsAppand(fmt.Sprintf("%s%s", dirPath, f.Name())) {
				manifest = append(manifest, scID[0])
			}
		}
	}
	return manifest
}

func CreateFile(content, file string) error {
	file = doFilepath(file)
	f, err := os.OpenFile(file, os.O_CREATE|os.O_RDWR, os.ModeAppend|os.ModePerm)
	if err != nil {
		l.Fatalf("fail to open file err=%v", err)
		return err
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	if err != nil {
		l.Fatalf("fail to write to file err=%v", err)
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
		if err := os.MkdirAll(outputPath, os.ModeDir|os.ModePerm); err != nil {
			l.Fatalf("%s create dir : %s", outputPath, err)
		}
	}
	count := 0
	for _, v := range filesName {
		yamlPath := fmt.Sprintf("%s/%s/manifest.yaml", outputPath, v)
		metaPath := fmt.Sprintf("%s/%s/meta.md", outputPath, v)
		id := strings.Split(v, "-")[0]
		md := Markdown{path: v, ID: id}
		_ = md.TemplateDecodeFile()
		if len(md.Description) < 1 {
			continue
		}
		yaml := Tmp{Path: "manifest.tpl", Obj: md, box: TemplateBox}
		meta := Tmp{Path: "md.tpl", Obj: md, box: TemplateBox}
		ScheckDir(v, outputPath)
		_ = CreateFile(yaml.GetTemplate(), yamlPath)
		_ = CreateFile(meta.GetTemplate(), metaPath)
		count++
	}

	l.Debugf("Template generation MF quantity =%d, in%s directory \n", count, outputPath)
}

// The parameter is a file list without a file suffix
func ToMakeMdFile(filesName []string, outputPath string) {
	if _, err := os.Stat(outputPath); err != nil {
		if err := os.MkdirAll(outputPath, os.ModeDir|os.ModePerm); err != nil {
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
		md := Markdown{path: v, ID: id}
		if err := md.TemplateDecodeFile(); err != nil {
			l.Fatalf("err:%s", err)
		}
		// Remove unwanted
		if len(md.Description) < 1 {
			continue
		}
		_, ok := category[md.Category]
		if ok {
			category[md.Category][fmt.Sprintf("%s-%s", strings.Split(md.RuleID, "-")[0], md.Title)] = md.RuleID
		}
		yuquemd := Tmp{Path: "yuquemd.tpl", Obj: md, box: TemplateBox}
		if _, err := os.Stat(outputPath); err != nil {
			if err := os.MkdirAll(outputPath, os.ModeDir|os.ModePerm); err != nil {
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

	yuque := Tmp{Path: "summary.tpl", Obj: Summary{Category: category}, box: TemplateBox}
	yuquePath := fmt.Sprintf("%s/%s", outputPath, "summary.md")

	err := ScheckDocSyncDisk(outputPath)
	if err != nil {
		fmt.Println(err)
	}

	_ = CreateFile(yuque.GetTemplate(), yuquePath)
}

/*
	Clear system script path
	Rewrite script

*/
func ScheckCoreSyncDisk(ruleDir string) {
	// Delete directory
	if _, err := os.Stat(ruleDir); err == nil {
		if err := os.RemoveAll(ruleDir); err != nil {
			l.Fatal(err)
		}
	}
	// Create a directory and synchronize Lua scripts to disk
	if _, err := os.Stat(ruleDir); err != nil {
		if err := os.Mkdir(ruleDir, os.ModeDir); err == nil {
			l.Debugf("The current scriptbox length is %d \n", len(ScriptBox.List()))
			for _, name := range ScriptBox.List() {
				if content, err := ScriptBox.Find(name); err == nil {
					name = strings.ReplaceAll(name, "\\", "/")
					// Processing multi-level directories
					paths := strings.Split(name, "/")
					if len(paths) > 1 {
						// Splice catalog
						libDir := fmt.Sprintf("%s/%s", ruleDir, strings.Join(paths[0:len(paths)-1], "/"))
						if _, err := os.Stat(libDir); err != nil {
							if err := os.MkdirAll(libDir, os.ModeDir|os.ModePerm); err != nil {
								l.Fatalf("%s create dir : %s", libDir, err)
							}
						}
					}
					_ = CreateFile(string(content), fmt.Sprintf("%s/%s", ruleDir, name))
				}
			}
		}
	}
}

func ScheckDocSyncDisk(filePath string) error {
	// Create a directory and synchronize Lua scripts to disk
	if _, err := os.Stat(filePath); err != nil {
		if err := os.Mkdir(filePath, os.ModeDir); err == nil {
			// Traversal Lua script name
			l.Debug("The current scriptbox length is %d \n", len(ScriptBox.List()))
		}
	}
	for _, name := range DocBox.List() {
		content, err := DocBox.Find(name)
		if err == nil {
			name = strings.ReplaceAll(name, "\\", "/")
			// Processing multi-level directories
			paths := strings.Split(name, "/")
			if len(paths) > 1 {
				libDir := fmt.Sprintf("%s/%s", filePath, strings.Join(paths[0:len(paths)-1], "/"))
				if _, err := os.Stat(libDir); err != nil {
					if err := os.MkdirAll(libDir, os.ModeDir|os.ModePerm); err != nil {
						l.Fatalf("%s create dir : %s", libDir, err)
					}
				}
			}
			res := fmt.Sprintf(string(content), git.Version, git.BuildAt)
			err := CreateFile(res, fmt.Sprintf("%s/%s", filePath, name))
			if err != nil {
				l.Fatalf("save file error,err :%s", err)
			}
		} else {
			l.Errorf("cannot find file %s from box", name)
		}
	}
	return nil
}
