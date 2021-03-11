# intro

目标：一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。

实际情况中存在两种恶意脚本，一种是恶意命令，如 `rm -rf`，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。

因此 sec_checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 sec_checker 将以客户端形式运行，并以日志方式通过统一的网络模型进行巡检事件的收集。同时 sec_checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

## `sec_checker` 架构（待定）

## `sec_checker` 规则脚本

规则脚本由两个部分组成

- `<rule-name>.lua`：这个是规则的判断脚本，基于lua语法实现，但不能引用，也无法引用标准lua库，只能使用内置的lua库和内置的函数。

脚本返回一个数据结构：

```python
{
	result: true/false,  # true/false 代表检测是否有问题

	# 以下是可扩展的返回字段，用于基于规则清单的模版变量
	var1: xxx,
	var2: yyy
}
```

- `<rule-name>.manifest`：这个是规则清单文件。当对应的 lua 脚本检测到有问题（result == true），manifest 文件中有一组对应的行为定义，这些字段如下：

```python
{
	rule_id: 事件的规则编号，如k8s-pod-001
	category: 事件的分类，如 security，network

	level: 当前事件的危险等级
	title: 当前事件的标题（支持模板）
	desc: 当前事件的内容（支持模板）
	cron: 配置事件的执行周期（就是 crontab 的语法规则）

	# 可扩展自定义的字段
	var1: xxx
	var2: yyy
	...
}

## sec_checker 内置函数

- `string file(path string)`: 读取文件内容，一般是文本文件
- `bool exist(path string)`: 判断路径文件是否存在
- `bool regex(regex, txt string)`: 判断 `txt` 是否匹配正则 `regex`
- `list iptables()`: 获取当前的 iptables 规则，返回规则对象列表
- `list processes()`: 获取当前所有进程列表（只需附带基础信息，如进程名，PID）
- `list processes()`: 获取当前所有进程列表（只需附带基础信息，如进程名，PID）
- `map processe_info(pid int)`: 获取当前进程 ID 的详情（尽可能拿全）
- `map file_info(path string)`: 获取当前文件的详情（尽可能拿全）
- `list ports()`: 获取打开的端口列表
- `list crontab()`: 获取系统上 crontab 列表
- `list login()`: 获取系统上当前登陆信息列表
- `list users()`: 获取系统上用户信息列表（不管当前是否登陆）
- `list groups()`: 获取系统上用户组信息列表（不管当前是否有登陆）
- `list shadow()`: 获取系统上 shadow 组信息列表
- `map kernel()`: 获取系统内核信息
- `map zone()`: 获取系统时区信息
- `date ntp()`：获取 ntp 时间
- `map ulimit()`: 获取系统 ulimit 信息
- `list mounts()`: 获取系统块设备列表
- `list eth()`: 获取系统网卡列表
- `string host()`: 获取系统主机名
- `list routs()`: 获取系统路由信息
- `list history()`: 获取系统历史命令列表
- `int uptime()`: 获取系统启动时长（返回单位为 s）
- `map ip_info(ip string)`: 获取 IP 详情
- `string date()`: 获取当前日期
- `map http_get(url string)`: 返回请求 url 的 HTTP 请求
- `string hash(path string)`: 获取文件的 md5 值
- `set_cache(key, val)`: 设置缓存
- `val get_cache(key)`: 获取缓存
- `val get_cache(key)`: 获取缓存
- `selinux()`: ?
- `sshd()`: ?
- `last()`: ?
- `list services()`: 获取当前系统上的服务列表
- `list python_packages()`: 获取当前系统上安装的 python 包情况（包名、包版本，python 版本）
