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
		{Name: `time_zone`, Fn: p.zone},
		{Name: `uptime`, Fn: p.uptime},
		{Name: `kernel_info`, Fn: p.kernelInfo},
		{Name: `kernel_modules`, Fn: p.kernelModules},
		{Name: `users`, Fn: p.users},
		{Name: `logged_in_users`, Fn: p.loggedInUsers},
		{Name: `last`, Fn: p.last},
		{Name: `lastb`, Fn: p.lastb},
		{Name: `shadow`, Fn: p.shadow},
		{Name: `ulimit_info`, Fn: p.ulimitInfo},
		{Name: `mounts`, Fn: p.mounts},
		{Name: `processes`, Fn: p.processes},
		{Name: `process_open_files`, Fn: p.processOpendFiles},
		{Name: `process_open_sockets`, Fn: p.processOpenSockets},
		{Name: `shell_history`, Fn: p.shellHistory},
		{Name: `iptables`, Fn: p.ipTables},
		{Name: `interface_addresses`, Fn: p.interfaceAddresses},
		{Name: `listening_ports`, Fn: p.listeningPorts},
		{Name: `crontab`, Fn: p.crontab},
		{Name: `uname`, Fn: p.uname},
		{Name: `sysctl`, Fn: p.sysctl},
		{Name: `rpm_list`, Fn: p.rpmList},
		{Name: `rpm_query`, Fn: p.rpmQuery},
	}

	return funcs
}

func (p *provider) Catalog() string {
	return "system"
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
