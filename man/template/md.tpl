
### 规则ID

{{ .RuleID }}


### 类别

{{ .Category }}


### 级别

{{ .Level }}


### 兼容版本
{{inc .OSArch }}



### 说明
{{range .Description}}
{{if . }}
> {{ . }}
{{else}}
无
{{end}}
{{end}}

### 扫描频率
{{ .Cron }}

### 理论基础

{{range .Rationale}}
{{if . }}
> {{ . }}
{{else}}
无
{{end}}
{{end}}




### 风险项
{{ inc .Riskitems }}


### 审计方法
{{ .Audit}}




### 补救
{{ .Remediation}}





### 影响
{{range .Impact}}
{{if . }}
{{ . }}
{{else}}
无
{{end}}
{{end}}


### 默认值
{{range .Defaultvalue}}
{{if . }}
{{ . }}
{{else}}
无
{{end}}
{{end}}


### 参考文献
{{htmlUrl .References}}


### CIS 控制
{{range .Cis}}
{{if . }}
{{ . }}
{{else}}
无
{{end}}
{{end}}
