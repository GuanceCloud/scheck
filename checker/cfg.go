package checker

import (
	"io/ioutil"

	"github.com/influxdata/toml"
)

const (
	MainConfigSample = `# ##(required) directory contains script
rule_dir='/usr/local/security-checker/rules.d'

# ##(required) output of the check result, support local file or remote http server
# ##localfile: file:///your/file/path
# ##remote:  http(s)://your.url
output='file:///var/log/security-checker/event.log'

log='/usr/local/security-checker/log'
log_level='info'	
`
)

var (
	Cfg *Config
)

type Config struct {
	RuleDir string `toml:rule_dir`
	Output  string `toml:output`

	DisableLog bool   `toml:disable_log`
	Log        string `toml:log`
	LogLevel   string `toml:log_level`
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
