package system

import "gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs"

type provider struct {
}

func (p *provider) Funcs() []funcs.Func {
	return []funcs.Func{
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
		{Name: `http_get`, Fn: p.httpGet},
		{Name: `listening_ports`, Fn: p.listeningPorts},
		{Name: `crontab`, Fn: p.crontab},
		{Name: `uname`, Fn: p.uname},
		{Name: `sysctl`, Fn: p.sysctl},
		{Name: `rpm_list`, Fn: p.rpmList},
		{Name: `rpm_query`, Fn: p.rpmQuery},
		{Name: `grep`, Fn: p.grep},
		{Name: `sc_path_watch`, Fn: p.pathWatch},
		{Name: `sc_sleep`, Fn: p.sleep},
		{Name: `sc_ticker`, Fn: p.ticker},
		{Name: `sc_log`, Fn: p.log},
	}
}

func (p *provider) Catalog() string {
	return "system"
}

func Init() {
	funcs.AddFuncProvider(&provider{})
}
