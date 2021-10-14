package config

import (
	"testing"

	bstoml "github.com/BurntSushi/toml"
)

var constr = `[system]
  rule_dir = "/usr/local/scheck/rules.d"
  custom_dir = "/usr/local/scheck/custom.rules.d"
  custom_rule_lib_dir = "/usr/local/scheck/custom.rules.d/libs"
  lua_HotUpdate = false
  lua_run_cap = 15
  lua_tot_cap = 20
  disable_log = false
  pprof = true

[scoutput]
  [scoutput.http]
    enable = true
    output = "http://127.0.0.1:9529/v1/write/security"
  [scoutput.log]
    enable = true
    output = "/var/log/scheck/event.log"
  [scoutput.alisls]
    enable = false
    endpoint = ""
    access_key_id = ""
    access_key_secret = ""
    project_name = "zhuyun-scheck"
    log_store_name = "scheck"

[logging]
  log = "/usr/local/scheck/log"
  log_level = "debug"
  rotate = 0

[cgroup]
  enable = false
  cpu_max = 10.0
  cpu_min = 5.0
  mem = 100
`

func TestTryLoadConfig(t *testing.T) {
	conf := &Config{}
	var err error

	_, err = bstoml.Decode(constr, conf)
	if err != nil {
		l.Errorf("unmarshal main cfg failed %s", err.Error())
		return
	}
	// 测试 配置文件比结构体多一个字段可以正常解析 结构比配置多一个字段 也可以正常解析
	t.Logf("decode config = %+v \n", conf.System)
	t.Logf("decode config = %+v \n", conf.ScOutput)
	t.Logf("decode config = %+v \n", conf.Logging)
	t.Logf("decode config = %+v \n", conf.Cgroup)

	conf1 := &Config{}
	bts := []byte(constr)
	err = bstoml.Unmarshal(bts, conf1)
	if err != nil {
		t.Logf("marshal err=%v", err)
		return
	}
	t.Logf("\n")
	t.Logf("decode config = %+v \n", conf1.System)
	t.Logf("decode config = %+v \n", conf1.ScOutput)
	t.Logf("decode config = %+v \n", conf1.Logging)
	t.Logf("decode config = %+v \n", conf1.Cgroup)
}
