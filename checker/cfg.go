package checker

import (
	"io/ioutil"

	"github.com/influxdata/toml"
)

var (
	Cfg *Config
)

type Config struct {
	Output  string `toml:output`
	RuleDir string `toml:rule_dir`
}

const (
	SampleMainConfig = `
output = ''
rule_dir = ''
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
