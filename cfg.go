package secChecker

import (
	"io/ioutil"

	"github.com/influxdata/toml"
)

var (
	Cfg *Config
)

type Config struct {
	Output    string `toml:output`
	RuleDir   string `toml:rule_dir`
	Log       string `toml:"log"`
	LogLevel  string `toml:"log_level"`
	LogRotate int    `toml:"log_rotate,omitempty"`
}

const (
	SampleMainConfig = `
output = ''
rule_dir = ''

log_level = 'info'
#log=''
#log_rotate = 32
`
)

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
