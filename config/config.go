package config

import (
	"io/ioutil"

	"github.com/influxdata/toml"
)

const (
	MainConfigSample = `# ##(required) directory contains script
rule_dir='/usr/local/scheck/rules.d'

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
	RuleDir string `toml:"rule_dir"`
	Output  string `toml:"output"`

	Cron string `toml:"cron"`

	DisableLog bool   `toml:"disable_log"`
	Log        string `toml:"log"`
	LogLevel   string `toml:"log_level"`
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
