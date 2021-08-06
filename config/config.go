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
	l   = logger.DefaultSLogger("config")
)

type Config struct {
	RuleDir       string `toml:"rule_dir"`
	CustomRuleDir string `toml:"custom"` // 用户自定义入口
	Output        string `toml:"output"`
	Cron          string `toml:"cron"`
	DisableLog    bool   `toml:"disable_log"`
	Log           string `toml:"log"`
	LogLevel      string `toml:"log_level"`

	Cgroup *Cgroup `toml:"cgroup"` // 动态控制cpu和Mem
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
