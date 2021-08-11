package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"

	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/logger"

	"github.com/influxdata/toml"
)

const (
	MainConfigSample = `[system]
  # ##(required) directory contains script
  # The system path, rules cannot be modified

  rule_dir='/usr/local/scheck/rules.d'

  # custom_rule_dir is a user-defined directory，You can modify the path

  custom_rule_dir='/usr/local/scheck/custom.rules.d'

  # ##(optional)global cron, default is every 10 seconds
  #cron='*/10 * * * *'
  #disable_log=false
[scoutput]
  # ##(required) output of the check result, support local file or remote http server or aliyun sls
  # ##remote:  http(s)://your.url
  [http]
     enable = true
     output='{{.Output}}'
  # ##localfile: file:///your/file/path
  [log]
     enable = false
     output='{{.Output}}'
  [alisls]
     enable = false
     endpoint = ''
     access_key_id = ''
     access_key_secret = ''
[logging]
  log='/usr/local/scheck/log'
  log_level='info'	

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
	CustomRuleDir string `toml:"custom"`        // 用户自定义入口
	LuaHotUpdate  string `toml:"lua_HotUpdate"` // lua热更开关
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
			//DisableLog:    true,
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
			AliSls: &AliSls{},
		},
		Logging: &Logging{
			LogLevel: "info",
			Log:      filepath.Join("/var/log/scheck", "log"),
		},
		Cgroup: &Cgroup{Enable: false, CPUMax: 30.0, CPUMin: 5.0, MEM: 50},
	}

	// windows 下，日志继续跟 datakit 放在一起
	if runtime.GOOS == "windows" {
		c.Logging.Log = filepath.Join(securityChecker.InstallDir, "log")
		c.System.RuleDir = filepath.Join(securityChecker.InstallDir, "rules.d")
		c.ScOutput.Log.Output = fmt.Sprintf("file://%s", filepath.Join(securityChecker.InstallDir, "event.log"))
	}

	return c
}

func LoadConfig(p string) error {
	cfgdata, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	c := &Config{}
	if err := toml.Unmarshal(cfgdata, c); err != nil {
		return err
	}
	Cfg = c

	return nil
}

// 查看当前的cpu和mem大小 控制cgroup的百分比 从而控制程序运行过程中占用系统资源的情况
func hostInfo() {
	cpuMun := runtime.NumCPU()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("当前cpu数量是%d 内存是%d", cpuMun, m.TotalAlloc)
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
