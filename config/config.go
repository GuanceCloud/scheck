package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"

	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

const (
	MainConfigSample = `[system]
  # ##(必选) 系统存放检测脚本的目录
  rule_dir = "{{.System.RuleDir}}"
  # ##客户自定义目录
  custom_dir = "{{.System.CustomRuleDir}}"
  # 可选 用户自定义lua库 不可用rule_dir 系统默认为用户目录下的libs
  custom_rule_lib_dir = "{{.System.CustomRuleLibDir}}"
  #热更新 默认false
  lua_HotUpdate = {{.System.LuaHotUpdate}}
  cron = "{{.System.Cron}}"
  #是否禁用日志
  disable_log = {{.System.DisableLog}}

[scoutput]
   # ##安全巡检过程中产生消息 可发送到本地、http、阿里云sls。
   # ##远程server，例：http(s)://your.url
  [scoutput.http]
    enable = {{.ScOutput.Http.Enable}}
    output = "{{.ScOutput.Http.Output}}"
  [scoutput.log]
    # ##可配置本地存储
    enable = {{.ScOutput.Log.Enable}}
    output = "{{.ScOutput.Log.Output}}"
  # 阿里云日志系统
  [scoutput.alisls]
    enable = {{.ScOutput.AliSls.Enable}}
    endpoint = "{{.ScOutput.AliSls.EndPoint}}"
    access_key_id = "{{.ScOutput.AliSls.AccessKeyID}}"
    access_key_secret = "{{.ScOutput.AliSls.AccessKeySecret}}"
    project_name = "{{.ScOutput.AliSls.ProjectName}}"
    log_store_name = "{{.ScOutput.AliSls.LogStoreName}}"

[logging]
  # ##(可选) 程序运行过程中产生的日志存储位置
  log = "{{.Logging.Log}}"
  log_level = "{{.Logging.LogLevel}}"
  rotate = {{.Logging.Rotate}}

[cgroup]
  # 可选 默认关闭 cpu是百分比
  enable = {{.Cgroup.Enable}}
  cpu_max = {{.Cgroup.CPUMax}}.0
  cpu_min = {{.Cgroup.CPUMin}}.0
  mem = {{.Cgroup.MEM}}
`
)

var (
	Cfg *Config
	l   = logger.DefaultSLogger("config")
)

type Config struct {
	System   *System   `toml:"system,omitempty"`
	ScOutput *ScOutput `toml:"scoutput"`
	Logging  *Logging  `toml:"logging,omitempty"`
	Cgroup   *Cgroup   `toml:"cgroup"`
}

type System struct {
	RuleDir          string `toml:"rule_dir"`
	CustomRuleDir    string `toml:"custom_dir"`
	CustomRuleLibDir string `toml:"custom_rule_lib_dir"`
	LuaHotUpdate     bool   `toml:"lua_HotUpdate"`
	Cron             string `toml:"cron"`
	DisableLog       bool   `toml:"disable_log"`
}

type ScOutput struct {
	Http   *Http   `toml:"http,omitempty"`
	Log    *Log    `toml:"log,omitempty"`
	AliSls *AliSls `toml:"alisls,omitempty"`
}

type Http struct {
	Enable bool   `toml:"enable"`
	Output string `toml:"output,omitempty"`
}
type Log struct {
	Enable bool   `toml:"enable"`
	Output string `toml:"output,omitempty"`
}
type AliSls struct {
	Enable          bool   `toml:"enable"`
	EndPoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	ProjectName     string `toml:"project_name"`
	LogStoreName    string `toml:"log_store_name"`
	Description     string `toml:"description,omitempty"`
	SecurityToken   string `toml:"security_token,omitempty"`
}

type Logging struct {
	Log      string `toml:"log"`
	LogLevel string `toml:"log_level"`
	Rotate   int    `toml:"rotate"`
}

// Cgroup cpu&mem 控制量
type Cgroup struct {
	Enable bool    `toml:"enable"`
	CPUMax float64 `toml:"cpu_max"`
	CPUMin float64 `toml:"cpu_min"`
	MEM    int     `toml:"mem"`
}

func DefaultConfig() *Config {

	c := &Config{
		System: &System{
			RuleDir:          "/usr/local/scheck/rules.d",
			CustomRuleDir:    "/usr/local/scheck/custom.rules.d",
			CustomRuleLibDir: "/usr/local/scheck/custom.rules.d/libs",
			Cron:             "",
			DisableLog:       false,
		},
		ScOutput: &ScOutput{
			Http: &Http{
				Enable: true,
				Output: "http://127.0.0.1:9529/v1/write/security",
			},
			Log: &Log{
				Enable: false,
				Output: fmt.Sprintf("file://%s", filepath.Join("/var/log/scheck", "event.log")),
			},
			AliSls: &AliSls{
				ProjectName:  "zhuyun-scheck",
				LogStoreName: "scheck",
			},
		},
		Logging: &Logging{
			LogLevel: "info",
			Log:      filepath.Join("/var/log/scheck", "log"),
			Rotate:   0, //默认32M
		},
		Cgroup: &Cgroup{Enable: false, CPUMax: 10, CPUMin: 5, MEM: 100},
	}

	// windows
	if runtime.GOOS == "windows" {
		c.Logging.Log = filepath.Join(securityChecker.InstallDir, "log")
		c.System.RuleDir = filepath.Join(securityChecker.InstallDir, "rules.d")
		c.System.CustomRuleDir = filepath.Join(securityChecker.InstallDir, "custom.rules.d")
		c.System.CustomRuleLibDir = filepath.Join(c.System.CustomRuleDir, "libs")
		c.ScOutput.Log.Output = fmt.Sprintf("file://%s", filepath.Join(securityChecker.InstallDir, "event.log"))
	}

	return c
}

// try load old config
func TryLoadConfig(filePath string) bool {
	type Conf struct {
		RuleDir       string `toml:"rule_dir"`
		CustomRuleDir string `toml:"custom_rule_dir"`
		Output        string `toml:"output"`
		Cron          string `toml:"cron"`
		Log           string `toml:"log"`
		LogLevel      string `toml:"log_level"`
		DisableLog    bool   `toml:"disable_log"`
	}
	oldConf := new(Conf)
	cfgData, err := ioutil.ReadFile(filePath)
	if err != nil {
		l.Fatalf("ReadFile err %v", err)
	}

	if err = toml.Unmarshal(cfgData, oldConf); err != nil {
		return true
	}
	newConf := DefaultConfig()
	if oldConf.CustomRuleDir != "" && oldConf.CustomRuleDir != newConf.System.CustomRuleDir {
		newConf.System.CustomRuleDir = oldConf.CustomRuleDir
	}
	if oldConf.Cron != "" {
		if oldConf.Cron != newConf.System.Cron {
			newConf.System.Cron = oldConf.Cron
		}
	}
	if oldConf.Log != "" {
		if oldConf.Log != newConf.Logging.Log {
			newConf.Logging.Log = oldConf.Log
		}
	}
	if oldConf.LogLevel != "" {
		if oldConf.LogLevel != newConf.Logging.LogLevel {
			newConf.Logging.LogLevel = oldConf.LogLevel
		}
	}
	tmplToFile(newConf, filePath)
	Cfg = newConf
	return false
}

func LoadConfig(p string) {
	cfgData, _ := ioutil.ReadFile(p)
	c := &Config{}
	fmt.Println(string(cfgData))
	if err := toml.Unmarshal(cfgData, c); err != nil {
		l.Fatalf("marshall  error:%v and  config is= %+v \n", err, c)
	}
	tmplToFile(c, p)
	Cfg = c
}

// Init config init
func (c *Config) Init() {
	// to init logging
	c.setLogging()
	//  CustomRuleDir file & rule.d file
	initDir()
}

//init log
func (c *Config) setLogging() {
	// set global log root
	if c.Logging.Log == "stdout" || c.Logging.Log == "" { // set log to disk file
		l.Info("set log to stdout")
		optFlags := logger.OPT_DEFAULT | logger.OPT_STDOUT
		if err := logger.InitRoot(
			&logger.Option{
				Level: c.Logging.LogLevel,
				Flags: optFlags}); err != nil {
			l.Errorf("set root log fatal: %s", err.Error())
		}
	} else {
		if c.Logging.Rotate > 0 {
			logger.MaxSize = c.Logging.Rotate
		}
		if err := logger.InitRoot(&logger.Option{
			Path:  c.Logging.Log,
			Level: c.Logging.LogLevel,
			Flags: logger.OPT_DEFAULT}); err != nil {
			l.Errorf("set root log faile: %s", err.Error())
		}
	}
	l = logger.DefaultSLogger("config")
}

// 初始化配置中的rule文件和用户自定义rules文件
func initDir() {
	_, err := os.Stat(Cfg.System.CustomRuleDir)
	if err != nil {
		_ = os.MkdirAll(Cfg.System.CustomRuleDir, 0644)
	}

	_, err = os.Stat(Cfg.System.RuleDir)
	if err != nil {
		_ = os.MkdirAll(Cfg.System.RuleDir, 0644)
	}

}

func hostInfo() {
	//cpuMun := runtime.NumCPU()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
}

func CreateSymlinks() {

	x := [][2]string{}

	if runtime.GOOS == "windows" {
		x = [][2]string{
			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck.exe"),
				`C:\WINDOWS\system32\scheck.exe`,
			},
		}
	} else {
		x = [][2]string{
			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck"),
				"/usr/local/bin/scheck",
			},

			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck"),
				"/usr/local/sbin/scheck",
			},

			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck"),
				"/sbin/scheck",
			},

			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck"),
				"/usr/sbin/scheck",
			},

			[2]string{
				filepath.Join(securityChecker.InstallDir, "scheck"),
				"/usr/bin/scheck",
			},
		}
	}

	for _, item := range x {
		if err := symlink(item[0], item[1]); err != nil {
			l.Warnf("create scheck symlink: %s -> %s: %s, ignored", item[1], item[0], err.Error())
		}
	}

}

func symlink(src, dst string) error {

	l.Debugf("remove link %s...", dst)
	if err := os.Remove(dst); err != nil {
		l.Warnf("%s, ignored", err)
	}

	return os.Symlink(src, dst)
}

func tmplToFile(c *Config, fpath string) {
	f, err := os.OpenFile(fpath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		l.Fatalf("open file err =%v", err)
	}
	defer f.Close()
	tmpl, err := template.New("config").Parse(MainConfigSample)
	if err != nil {
		l.Fatalf("make template err=%v", err)
	}
	if runtime.GOOS == "windows" {
		c.System.RuleDir = strings.ReplaceAll(c.System.RuleDir, "\\", "\\\\")
		c.System.CustomRuleLibDir = strings.ReplaceAll(c.System.CustomRuleLibDir, "\\", "\\\\")
		c.System.CustomRuleDir = strings.ReplaceAll(c.System.CustomRuleDir, "\\", "\\\\")
		c.Logging.Log = strings.ReplaceAll(c.Logging.Log, "\\", "\\\\")
		c.ScOutput.Log.Output = strings.ReplaceAll(c.ScOutput.Log.Output, "\\", "\\\\")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, c)
	if err != nil {
		l.Fatalf("execute err=%v", err)
	}
	//_, err = f.WriteString(buf.String())
	_, err = f.Write(buf.Bytes())
	if err != nil {
		l.Fatalf("err:=%v", err)
	}
}
