package utils

import securityChecker "gitlab.jiagouyun.com/cloudcare-tools/sec-checker"

type provider struct {
	funcs []securityChecker.Func
}

func (p *provider) Funcs() []securityChecker.Func {
	funcs := []securityChecker.Func{
		{Name: `set_cache`, Fn: p.setCache},
		{Name: `get_cache`, Fn: p.getCache},
		{Name: `json_encode`, Fn: p.jsonEncode},
		{Name: `json_decode`, Fn: p.jsonDecode},
	}
	return funcs
}

func (p *provider) Catalog() string {
	return "utils"
}

func init() {
	securityChecker.AddFuncProvider(&provider{})
}
