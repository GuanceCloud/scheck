package system

import (
	securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"
)

type provider struct {
	funcs []securityChecker.Func
}

func (p *provider) Funcs() []securityChecker.Func {
	funcs := []securityChecker.Func{
		{Name: `ls`, Fn: p.ls},
		{Name: `file_exist`, Fn: p.fileExist},
		{Name: `file_info`, Fn: p.fileInfo},
		{Name: `read_file`, Fn: p.readFile},
		{Name: `file_hash`, Fn: p.fileHash},
		{Name: `hostname`, Fn: p.hostname},
		{Name: `zone`, Fn: p.zone},
		{Name: `uptime`, Fn: p.uptime},
		{Name: `kernel_info`, Fn: p.kernelInfo},
		{Name: `kernel_modules`, Fn: p.kernelModules},
		{Name: `users`, Fn: p.users},
		{Name: `shadow`, Fn: p.shadow},
		{Name: `ulimit_info`, Fn: p.ulimitInfo},
		{Name: `mounts`, Fn: p.mounts},
		{Name: `processes`, Fn: p.processes},
		{Name: `shell_history`, Fn: p.shellHistory},
		{Name: `last`, Fn: p.last},
		{Name: `iptables`, Fn: p.ipTables},
		{Name: `interface_addresses`, Fn: p.interfaceAddresses},
		{Name: `lasprocess_open_socketst`, Fn: p.processOpenSockets},
		{Name: `listening_ports`, Fn: p.listeningPorts},
	}

	return funcs
}

func (p *provider) Catalog() string {
	return "system"
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
