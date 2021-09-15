# Scheck 入门简介


- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64


本文档主要介绍 [Scheck 安装](scheck-install)完后，如何使用 Scheck 中的基本功能，包括如下几个方面：

- 安装介绍

## Scheck 目录介绍

Scheck 目前支持 Linux/Windows 两种主流平台：


| 操作系统                            | 架构                | 安装路径                                                                                     |
| ---------                           | ---                 | ------                                                                                       |
| Linux 内核 2.6.23 或更高版本        | amd64/386/arm/arm64 | `/usr/local/scheck`                                                      |
| Windows 7, Server 2008R2 或更高版本 | amd64/386           | 64位：`C:\Program Files\scheck`<br />32位：`C:\Program Files(32)\scheck` |

> Tips：查看内核版本

- Linux：`uname -r`
- Windows：执行 `cmd` 命令（按住 Win键 + `r`，输入 `cmd` 回车），输入 `winver` 即可获取系统版本信息

安装完成后，Scheck 目录列表大概如下：

```
├── [   6]  custom.rules.d
├── [ 12K]  rules.d
├── [ 17M]  scheck
├── [ 664]  scheck.conf
└── [ 222]  version
```

其中：

- `scheck`：Scheck 主程序，Windows 下为 `scheck.exe`
- `custom.rules.d`：用户自定义目录
- `rules.d`：Scheck 系统目录
- `scheck.conf` Scheck 主配置文件
- `version`：Scheck 版本信息

> 注：Linux 平台下，Scheck 运行日志在 `/var/log/scheck` 目录下。


## 检测规则

检测规则放在规则目录中：由配置文件中 `rule_dir` 或是自定义用户目录`custom_dir`指定。每项规则对应两个文件：  

1. 脚本文件：使用 [Lua](http://www.lua.org/) 语言来编写，必须以 `.lua` 为后缀名。    
2. 清单文件：使用 [TOML](https://toml.io/en/) 格式，必须以 `.manifest` 为后缀名，[详情](#清单文件)。  

脚本文件和清单文件**必须同名**。

Security Checker 会定时周期性(由清单文件的 `cron` 指定)地执行检测脚本，Lua 脚本代码每次执行时检测相关安全事件(比如文件被改动、有新用户登录等)是否触发，如果触发了，则使用 `trigger()` 函数将事件(以行协议格式)发送到由配置文件中 `output` 字段指定的地址。   

Security Checker 定义了若干 Lua 扩展函数，并且为确保安全，禁用了一些 Lua 包或函数，仅支持以下 Lua 内置包/函数：  

- 以下内置基础包均支持
	- `table`
	- `math`
	- `string`
	- `debug`

- `os` 包中，**除以下函数**外，均可使用：
	- `execute()`
	- `remove()`
	- `rename()`
	- `setenv()`
	- `setlocale()`

> 添加/修改规则清单文件以及 Lua 代码，**不需要重启服务**，Security Checker 会每隔 10 秒扫描规则目录。

### 清单文件 

清单文件是对当前规则所检测内容的一个描述，比如检测文件变化，端口启停等。最终的行协议数据中只会包含清单文件中的字段。详细内容如下：    

```toml
# ---------------- 必选字段 ---------------

# 事件的规则编号，如 k8s-pod-001，将作为行协议的指标名
id = ''

# 事件的分类，根据业务自定义
category = ''

# 当前事件的危险等级，根据业务自定义，比如warn，error
level = ''

# 当前事件的标题，描述检测内容，如 "敏感文件改动"
title = ''

# 当前事件的内容（支持模板，详情见下）
desc = ''

# 配置事件的执行周期（使用 linux crontab 的语法规则）
cron = ''

# 平台支持
os_arch = ["Linux", "Windows"]
# ---------------- 可选字段 ---------------

# 禁用当前规则
#disabled = false

# 默认在tag中添加hostname
#omit_hostname = false

# 显式设置hostname
#hostname = ''

# ---------------- 自定义字段 ---------------

# 支持添加自定义key-value，且value必须为字符串
#instanceID=''
```
### Cron规则
```
# ┌───────────── Second
# │ ┌───────────── Miniute
# │ │ ┌───────────── Hour
# │ │ │ ┌───────────── Day-of-Month
# │ │ │ │ ┌───────────── Month
# │ │ │ │ │                                   
# │ │ │ │ │
# │ │ │ │ │
# * * * * *
```

例：  
`10 * * * *`: 在每分钟的第 10 秒运行。  
`*/10 * * * *`: 每隔 10 秒运行。  
`10 1 * * *`: 在每小时的 1 分 10 秒运行。  
`10 */3 * * *`: 在每 3 分钟的第 10 秒运行，比如：01:03:10, 01:06:10 ...  


### 模板支持

清单文件中 `desc` 的字符串中支持设置模板变量，语法为 `{{.<Variable>}}`，比如

```
文件{{.FileName}}被修改，改动后的内容为: {{.Content}}
```

表示 `FileName` 和 `Diff` 是模板变量，将会被替换(包括前面的点号 `.`)。调用 `trigger()` 函数时，变量即实现了替换。 该函数可传入一个 Lua 的 `table` 变量，其中包含了模板变量的替换值，假设传入如下参数：  

```lua
tmpl_vals={
    FileName = "/etc/passwd",
    Content = "delete user demo"
}
trigger(tmpl_vals)
```

则最终的 `desc` 值为：

```
文件/etc/passwd被修改，改动内容为: delete user demo
```

## 测试规则

在编写规则代码的时候，可以使用`scheck --test`来测试代码是否正确。  
假设 rules.d 目录下有一个 demo 规则：    

```shell
$ scheck --test  ./rules.d/demo
```

## lua 函数

见 [函数](funcs)

## 创建通用库

Security Checker 允许在检测脚本中使用 `require` 函数来导入 Lua 模块，模块文件只能存放在 `rules.d/lib` 目录下。可以将一些常用的函数模块化，放在这个 lib 子目录下供检测脚本使用。  

假设创建一个Lua模块 `common.lua`：  

``` lua
module={}

function modules.Foo()
    -- 函数体...
end

return module
```

将 `common.lua` 放在 `/usr/local/scheck/rules.d/lib` 目录下。

假设有规则脚本 demo.lua 使用该通用模块：  

``` lua
common=require("common") --不需要写后缀名
common.Foo()
```

## 行协议

Scheck 的输出为行协议格式。以规则 ID 为指标名。

### 标签列表（tags）

| Name       | Type   | Description                                    | Required |
| ---        |:----:  | ----                                           | :---:    |
| `title`    | string | 安全事件标题                                   | true     |
| `category` | string | 事件分类                                       | true     |
| `level`    | string | 安全事件等级，支持：`info`，`warn`，`critical` | true     |
| `host`     | string | 事件来源主机名（默认有） 
| `os_arch`  | string | 主机平台                                    | true    |
| 自定义tags | string | 在清单文件中自定义的tag                        | false    |

目前的几种 `category` 分类

- `network`: 网络相关，主要涉及连接、端口、防火墙等
- `storage`：存储相关，如磁盘、HDFS 等
- `db`：各种数据库相关（MySQL/Redis/...）
- `system`：主要是操作系统相关
- `container`：包括 Docker 和 Kubernetes

### 指标列表(fields)

| 指标名    | 类型   | 描述     |
| ---       | :---:  | ----     |
| `message` | string | 事件详情 |

### Scheck 限制运行资源
  通过 cgourp 限制Scheck 运行资源（例如 CPU 使用率等），仅支持 Linux 操作系统。
  进入 限制Scheck 安装目录下，修改 scheck.conf 配置文件，将 enable 设置为 true，示例如下：
 ```toml
[cgroup]
    # 可选 默认关闭 可控制cpu和mem
  enable = false
  cpu_max = 30.0
  cpu_min = 5.0
  mem = 0
 ```
