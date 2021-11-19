# Summary

- [Scheck 是什么](scheck-info)
- [Scheck 版本历史](changelog)
- [Scheck 使用入门](scheck-how-to)
- [Scheck 最佳实践](best-practices)
- [Scheck 安装、配置]()
	- [Scheck 安装](scheck-install)
	- [Scheck 配置](scheck-configure)
- [Scheck 多端输出]()
	- [Scheck 连接Datakit](join-datakit)
	- [Scheck 连接阿里云日志](join-sls)

- [Scheck 脚本二次开发]()
	- [检查敏感文件的变动实现](scheck-filechange)
	- [监控系统用户的变化](scheck-userchange)
	- [用户自定义规则及lib库](custom-how-to)

- [lua内置标准库和lua-lib](lualib)
- [Scheck Funcs支持清单](funcs)

- [其他]()
    - [Scheck 并发策略](scheck-pool)

- [规则库]()
{{ range $index, $value := .Category }}
    - [{{$index -}}]()
   {{range $index, $value := $value }}
        - [{{ $index}}]({{$value}})
   {{ end -}}
{{ end }}
