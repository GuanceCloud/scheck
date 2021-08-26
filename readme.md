# Security Checker

一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。  
实际情况中存在两种恶意脚本，一种是恶意命令，如 `rm -rf`，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。 
因此 Security Checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 Security Checker 将以日志方式通过统一的网络模型进行巡检事件的收集。同时 Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。


[安装/更新](#安装/更新)  
[配置](#配置)  
[规则](#检测规则)  
[测试规则](#测试规则)  
[行协议](#行协议)  
[函数](./funcs.md)  
[示例](#示例)  

## 安装/更新
### Linux 平台
*安装*：  
```Shell
sudo -- sh -c 'curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/scheck/install.sh'
```

*更新*：  
```Shell
SC_UPGRADE=true;sudo -- sh -c 'curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/scheck/install.sh'
```

安装完成后即以服务的方式运行，服务名为`scheck`，使用服务管理工具来控制程序的启动/停止：  

```
systemctl start/stop/restart scheck
```

或

```
service scheck start/stop/restart
```
### Windows平台
*安装*：
```powershell
Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/scheck/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

*更新*：
```powershell
$env:SC_UPGRADE;Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/scheck/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

> 注意：Security Checker 支持 Linux/Windows amd64/arm64


可配置环境变量`scoutput`来设置 Security Checker 安装后的初始 output。  

默认安装目录为 `/usr/local/scheck`。
安装完成后会将可用系统规则安装到`rules.d`子目录下。Security Checker的更新会更新规则包。

## 配置

进入默认安装目录 `/usr/local/scheck`，打开配置文件 `scheck.conf`，配置文件采用 [TOML](https://toml.io/en/) 格式，说明如下：

```toml
[system]
  # ##(必选) 系统存放检测脚本的目录
  rule_dir = "/usr/local/scheck/rules.d"
  # ##客户自定义目录
  custom_dir = "/usr/local/scheck/custom.rules.d"
  #热更新
  lua_HotUpdate = ""
  cron = ""
  #是否禁用日志
  disable_log = false

[scoutput]
   # ##安全巡检过程中产生消息 可发送到本地、http、阿里云sls。
   # ##远程server，例：http(s)://your.url
  [scoutput.http]
    enable = true
    output = "http://127.0.0.1:9529/v1/write/security"
  [scoutput.log]
    # ##可配置本地存储
    enable = false
    output = "/var/log/scheck/event.log"
  # 阿里云日志系统
  [scoutput.alisls]
    enable = false
    endpoint = ""
    access_key_id = ""
    access_key_secret = ""
    project_name = "zhuyun-scheck"
    log_store_name = "scheck"

[logging]
  # ##(可选) 程序运行过程中产生的日志存储位置
  log = "/var/log/scheck/log"
  log_level = "info"
  rotate = 0

[cgroup]
    # 可选 默认关闭 可控制cpu和mem
  enable = false
  cpu_max = 30.0
  cpu_min = 5.0
  mem = 0

```

> **主配置文件**的修改**需要重启服务**才生效

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

见 [函数](./funcs.md)

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

Security Checker 的输出为行协议格式。以规则 ID 为指标名。

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

## 检测示例

### 检查敏感文件的变动

一旦敏感文件有变动，则将事件记录到文件 `/var/log/scheck/event.log` 中。    

1. 进入安装目录，编辑配置文件 `scheck.conf` 的 `output` 字段：  

```toml
rule_dir  = '/usr/local/scheck/rules.d'
output    = '/var/log/scheck/event.log'
log       = '/usr/local/scheck/log'
log_level = 'info'
```

2. 在目录 `/usr/local/scheck/rules.d`(即以上配置文件中的 `rule_dir`) 下新建清单文件 `files.manifest`，编辑如下：  

```toml
id       = 'check-file'
category = 'system'
level    = 'warn'
title    = '监视文件变动'
desc     = '文件 {{.File}} 发生了变化'
cron     = '*/10 * * * *' #表示每10秒执行该lua脚本
os_arch  = ["CentOS", "Darwin"]
```

3. 在清单文件同级目录下新建脚本文件 `files.lua`，编辑如下：

```lua
local files={
	'/etc/passwd',
	'/etc/group'
}

local function check(file)
	local cache_key=file
	local hashval = file_hash(file)

	local old = get_cache(cache_key)
	if not old then
		set_cache(cache_key, hashval)
		return
	end

	if old ~= hashval then
		trigger({File=file})
		set_cache(cache_key, hashval)
	end
end

for i,v in ipairs(files) do
	check(v)
end
```

4. 当有敏感文件被改动后，下一个 10 秒会检测到并触发 trigger 函数，从而将事件发送到文件 `/var/log/scheck/event.log` 中，添加一行数据，例：  

```
check-file-01,category=security,level=warn,title=监视文件变动 message="文件 /etc/passwd 发生了变化" 1617262230001916515
```

### 监控系统用户的变化。    

基础步骤上上面类似，以下是检测规则代码和清单文件。  

规则文件

```toml
id         = 'users-checker'
category   = 'system'
level      = 'warn'
title      = '监视系统用户变动'
desc       = '{{.Content}}'
cron       = '*/10 * * * *'
instanceId = 'id-xxx'
os_arch    = ["Linux"]
```

检测代码 `users.lua`：  

```lua
local function check()
    local cache_key="current_users"
    local currents=users()

    local old=get_cache(cache_key)
    if not old then
        set_cache(cache_key, currents)
        return
    end

    local adds={}
    for i,v in ipairs(currents) do
        local exist=false
        for ii,vv in ipairs(old) do
            if vv["username"] == v["username"] then
                exist = true
                break
            end
        end
        if not exist then
            table.insert(adds, v["username"])
        end
    end

    local dels={}
    for i,v in ipairs(old) do
        local exist=false
        for ii,vv in ipairs(currents) do
            if vv["username"] == v["username"] then
                exist = true
                break
            end
        end
        if not exist then
            table.insert(dels, v["username"])
        end
    end

    local content=''
    if #adds > 0 then
        content=content..'新用户: '..table.concat(adds, ',')
    end
    if #dels > 0 then
        if content ~= '' then content=content..'; ' end
        content=content..'删除的用户: '..table.concat(dels, ',')
    end
    if content ~= '' then
        trigger({Content=content})
        set_cache(cache_key, currents)
    end
end

check()
```

### 监控是否有新的端口开启

一旦有新的端口被绑定，则将安全事件发送到 DataFlux

1. 将配置文件 `scheck.conf` 的 `output` 设置为 DataKit 的 server 地址：  

```toml
rule_dir  = '/usr/local/scheck/rules.d'
output    = 'http://localhost:9529/v1/write/security' # datakit 1.1.6(含)以上版本才支持
log       = '/usr/local/scheck/log'
log_level = 'info'
```

2. 在目录 `/usr/local/scheck/rules.d` (即以上配置文件中的 `rule_dir`) 下新建清单文件 `ports.manifest`，编辑如下：  

```toml
id       = 'check-listening-ports'
category = 'network'
level    = 'warn'
title    = '监视新端口'
desc     = '{{.Content}}'
cron     = '*/10 * * * *'
os_arch   = ["CentOS", "Darwin"]
```

3. 在清单文件同级目录下新建脚本文件 `ports.lua`，编辑如下：

```lua
local function check()
    local cache_key='check_ports'
    local ports = listening_ports()
    local old=get_cache(cache_key)
    if not old then
        set_cache(cache_key, ports)
        return
    end

    local content=''
    for i,v in ipairs(ports) do
        if v["family"] == 'AF_INET' then
            local exist=false
            for ii,vv in ipairs(old) do
                if vv["port"] == v["port"] and vv["address"] == v["address"]  then
                    exist = true
                    break
                end
            end
            if not exist then
                if content ~= '' then content=content.."; " end
                content = content..string.format("%d(%s)", v["port"], v["protocol"])
            end
        end
    end

    if content ~= '' then
        trigger({Content="发现新端口: "..content})
        set_cache(cache_key, ports)
    end
end

check()
```
