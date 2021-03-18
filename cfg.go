package secChecker

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/influxdata/toml"
)

var (
	Cfg *Config
)

func DefaultConfig(installDir string) *Config {
	return &Config{
		Output:    "",
		LogLevel:  "info",
		Log:       filepath.Join(installDir, "log"),
		LogRotate: 32,
		LogUpload: false,
	}
}

type Config struct {
	Output    string `toml:output`
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

	c := &Config{}
	if err := toml.Unmarshal(cfgdata, c); err != nil {
		return err
	}
	Cfg = c

	return nil
}

func (c *Config) Dump(cfgpath string) error {

	if mcdata, err := toml.Marshal(c); err != nil {
		l.Errorf("Toml Marshal(): %s", err.Error())
		return err
	} else {

		if err := ioutil.WriteFile(cfgpath, mcdata, 0644); err != nil {
			l.Errorf("error creating %s: %s", cfgpath, err)
			return err
		}
	}

	return nil
}

func createDefaultConfigFile(installDir string) error {
	// build default main config when install
	cfg := DefaultConfig(installDir)
	if err := cfg.Dump(MainConfPath(installDir)); err != nil {
		return fmt.Errorf("failed to init main config: %s", err.Error())
	}
	return nil
}
