package ssh

import (
	lua "github.com/yuin/gopher-lua"
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

type provider struct {
	funcs []securityChecker.Func
}

func (p *provider) sshConfigs(l *lua.LState) int {

	return 1
}

func (p *provider) Funcs() []securityChecker.Func {
	funcs := []securityChecker.Func{
		{Name: `ssh_configs`, Fn: p.sshConfigs},
	}
	return funcs
}

func (p *provider) Catalog() string {
	return "ssh"
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
