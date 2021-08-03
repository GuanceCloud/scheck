package config

import (
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/influxdata/toml"
)

const (
	MainConfigSample = `# ##(required) directory contains script
rule_dir='/usr/local/scheck/rules.d'
custom_rule_dir='/usr/local/scheck/custom.rules.d'
# ##(required) output of the check result, support local file or remote http server
# ##localfile: file:///your/file/path
# ##remote:  http(s)://your.url
output='{{.Output}}'


# ##(optional)global cron, default is every 10 seconds
#cron='*/10 * * * *'

log='/usr/local/scheck/log'
log_level='info'	
#disable_log=false
`
)

var (
	Cfg *Config
)

type Config struct {
	RuleDir       string `toml:"rule_dir"`
	CustomRuleDir string `toml:"custom"` // 用户自定义入口
	Output        string `toml:"output"`
	Cron          string `toml:"cron"`
	DisableLog    bool   `toml:"disable_log"`
	Log           string `toml:"log"`
	LogLevel      string `toml:"log_level"`

	Cgroup *Cgroup `toml:"cgroup"` // 动态控制
}

// Cgroup cpu&mem 控制量
type Cgroup struct {
	Enable bool    `toml:"enable"`
	CPUMax float64 `toml:"cpu_max"`
	CPUMin float64 `toml:"cpu_min"`
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
