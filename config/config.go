package config

import (
	"fmt"
	"io/ioutil"
	"runtime"

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
)

type Config struct {
	RuleDir       string `toml:"rule_dir,omitempty"`
	CustomRuleDir string `toml:"custom_rule_dir,omitempty"` // 用户自定义入口
	Output        string `toml:"output,omitempty"`
	Cron          string `toml:"cron,omitempty"`
	DisableLog    bool   `toml:"disable_log,omitempty"`
	Log           string `toml:"log,omitempty"`
	LogLevel      string `toml:"log_level,omitempty"`
	System        *System `toml:"system,omitempty"`
	ScOutput      *ScOutput `toml:"scoutput"`
	Logging       *Logging `toml:"logging,omitempty"` // 日志配置
	Cgroup *Cgroup `toml:"cgroup"` // 动态控制
}

type System struct {
	RuleDir       string `toml:"rule_dir"`
	CustomRuleDir string `toml:"custom"` // 用户自定义入口
	Cron          string `toml:"cron"`
	DisableLog    bool   `toml:"disable_log"`
}

type ScOutput struct {
	Http *Http `toml:"output,omitempty"`
	Log *Log  `toml:"log,omitempty"`
	AliSls *AliSls `toml:"alisls,omitempty"`

}

type Http struct {
	Enable bool    `toml:"enable"`
	Output        string `toml:"output,omitempty"`
}
type Log struct {
	Enable bool    `toml:"enable"`
	Output        string `toml:"output,omitempty"`
}
type AliSls struct {
	Enable bool    `toml:"enable"`
	EndPoint        string `toml:"endpoint"`
	AccessKeyID     string `toml:"access_key_id"`
	AccessKeySecret string `toml:"access_key_secret"`
	ProjectName  string `toml:"project_name"`
	LogStoreName  string `toml:"log_store_name"`
	Description  string `toml:"description,omitempty"`
}

type Logging struct {
	Log           string `toml:"log"`
	LogLevel      string `toml:"log_level"`
}

// Cgroup cpu&mem 控制量
type Cgroup struct {
	Enable bool    `toml:"enable"`
	CPUMax float64 `toml:"cpu_max"`
	CPUMin float64 `toml:"cpu_min"`
}

func DefaultConfig() *Config {
	c := &Config{
		System: &System{RuleDir: "",
			CustomRuleDir: "",
			Cron:          "",
			DisableLog:    true,
		},
		ScOutput: &ScOutput{
			Http: &Http{
				Enable: true,
				Output: "",
			},
			Log: &Log{
				Enable: false,
				Output: "",
			},
			AliSls: &AliSls{},
		},

		Cgroup: &Cgroup{Enable: false, CPUMax: 30.0, CPUMin: 5.0},
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
