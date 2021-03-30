package checker

import (
	"io/ioutil"

	"github.com/influxdata/toml"
)

var (
	Cfg *Config
)

type Config struct {
	RuleDir string `toml:rule_dir`
	Output  string `toml:output`

	Log      string `toml:log`
	LogLevel string `toml:log_level`
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
