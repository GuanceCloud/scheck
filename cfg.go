package secChecker

import (
	"io/ioutil"
	"path/filepath"

	"github.com/influxdata/toml"
)

var (
	Cfg *Config
)

func DefaultConfig() *Config {
	return &Config{
		LogLevel:  "info",
		Log:       filepath.Join(InstallDir, "log"),
		LogRotate: 32,
		LogUpload: false,
	}
}

type Config struct {
	Log       string `toml:"log"`
	LogLevel  string `toml:"log_level"`
	LogRotate int    `toml:"log_rotate,omitempty"`
	LogUpload bool   `toml:"log_upload"`
}

func LoadConfig(p string) error {
	cfgdata, err := ioutil.ReadFile(p)
	if err != nil {
		return err
	}

	c := DefaultConfig()
	if err := toml.Unmarshal(cfgdata, c); err != nil {
		return err
	}
	Cfg = c

	return nil
}

func (c *Config) InitCfg(p string) error {

	if mcdata, err := toml.Marshal(c); err != nil {
		l.Errorf("Toml Marshal(): %s", err.Error())
		return err
	} else {

		if err := ioutil.WriteFile(p, mcdata, 0644); err != nil {
			l.Errorf("error creating %s: %s", p, err)
			return err
		}
	}

	return nil
}
