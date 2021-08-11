package utils

import securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"

type provider struct {
	funcs []securityChecker.Func
}

func (p *provider) Funcs() []securityChecker.Func {
	funcs := []securityChecker.Func{
		{Name: `set_cache`, Fn: p.setCache},
		{Name: `get_cache`, Fn: p.getCache},
		{Name: `set_global_cache`, Fn: p.setGlobalCache},
		{Name: `get_global_cache`, Fn: p.getGlobalCache},
		{Name: `json_encode`, Fn: p.jsonEncode},
		{Name: `json_decode`, Fn: p.jsonDecode},
		{Name: `mysql_weak_psw`, Fn: p.checkMysqlWeakPassword},
		{Name: `mysql_ports_list`, Fn: p.mysqlPortsList},
	}
	return funcs
}

func (p *provider) Catalog() string {
	return "utils"
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
