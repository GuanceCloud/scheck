package config

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/influxdata/toml"
	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

const (
	// todo 覆盖写配置文件时 要想带上注释 必须用模版。。。
	MainConfigSample = `[system]
  # ##(必选) 系统存放检测脚本的目录
  rule_dir = "/usr/local/scheck/rules.d"
  # ##客户自定义目录
  custom_dir = "/usr/local/scheck/custom.rules.d"
  #热更新
  lua_HotUpdate = false
  cron = ""
  #是否禁用日志
  disable_log = false

[scoutput]
    # ##安全巡检过程中产生消息 可发送到本地、http、阿里云sls。
   # ##远程server，例：http(s)://your.url
  [scoutput.http]
    enable = true
    output = "http://127.0.0.1:9529/v1/write/security"
  [scoutput.log]
    # ##可配置本地存储
    enable = false
    output = "/var/log/scheck/event.log"
  # 阿里云日志系统
  [scoutput.alisls]
    enable = false
    endpoint = ""
    access_key_id = ""
    access_key_secret = ""
    project_name = "zhuyun-scheck"
    log_store_name = "scheck"

[logging]
  # ##(可选) 程序运行过程中产生的日志存储位置
  log = "/var/log/scheck/log"
  log_level = "info"
  rotate = 0

[cgroup]
    # 可选 默认关闭 可控制cpu和mem
  enable = false
  cpu_max = 30.0
  cpu_min = 5.0
  mem = 0
`
)

var (
	Cfg *Config
	l   = logger.DefaultSLogger("config")
)

type Config struct {
	System   *System   `toml:"system,omitempty"`
	ScOutput *ScOutput `toml:"scoutput"`
	Logging  *Logging  `toml:"logging,omitempty"` // 日志配置
	Cgroup   *Cgroup   `toml:"cgroup"`            // 动态控制
}

type System struct {
	RuleDir       string `toml:"rule_dir"`
	CustomRuleDir string `toml:"custom_dir"`    // 用户自定义入口
	LuaHotUpdate  bool   `toml:"lua_HotUpdate"` // lua热更开关
	Cron          string `toml:"cron"`
	DisableLog    bool   `toml:"disable_log"`
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
	Log      string  `toml:"log"`
	LogLevel string  `toml:"log_level"`
	Cgroup   *Cgroup `toml:"cgroup"` // 动态控制cpu和Mem
	Rotate   int     `toml:"rotate"` // 日志分片大小 单位M 默认30M
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
			RuleDir:       "/usr/local/scheck/rules.d",
			CustomRuleDir: "/usr/local/scheck/custom.rules.d",
			Cron:          "",
			DisableLog:    false,
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
		Cgroup: &Cgroup{Enable: false, CPUMax: 30.0, CPUMin: 5.0, MEM: 200},
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == "windows" {
		c.Logging.Log = filepath.Join(securityChecker.InstallDir, "log")
		c.System.RuleDir = filepath.Join(securityChecker.InstallDir, "rules.d")
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
		log.Fatalf("ReadFile err %v", err)
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
	// 将新的配置写到配置文件中 O_TRUNC 覆盖写
	f, _ := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	res, _ := securityChecker.TomlMarshal(newConf)
	// bug:将bytes写到文件时 没有格式化。转换成string后就格式化
	_, err = f.WriteString(string(res))
	if err != nil {
		log.Fatalf("err:=%v", err)
	}

	Cfg = newConf
	return false
}

// LoadConfig
func LoadConfig(p string) {
	cfgData, _ := ioutil.ReadFile(p)
	c := &Config{}
	if err := toml.Unmarshal(cfgData, c); err != nil {
		log.Fatalf("marshall  error:%v and  config is= %+v \n", err, c)
	}
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

// 查看当前的cpu和mem大小 控制cgroup的百分比 从而控制程序运行过程中占用系统资源的情况
func hostInfo() {
	//cpuMun := runtime.NumCPU()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	//fmt.Printf("当前cpu数量是%d 内存是%d", cpuMun, m.TotalAlloc)
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
