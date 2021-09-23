# 检查敏感文件的变动实现
  本次将展示如何使用Scheck 检查敏感文件的lua脚本实现。

- 版本：%s
- 发布日期：%s
- 操作系统支持：linux/arm,linux/arm64,linux/386,linux/amd64  
 
## 前提条件

- 已安装[Scheck](scheck-install)

## 开发步骤

1. 进入安装目录，编辑配置文件 `scheck.conf` 的 `enable` 字段 设置为`true`：  

```toml
...
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
...
```

2. 在目录 `/usr/local/scheck/custom.rules.d`(此目录为用户自定义脚本目录) 下新建清单文件 `files.manifest`，编辑如下：  

```toml
id       = 'check-file'
category = 'system'
level    = 'warn'
title    = '监视文件变动'
desc     = '文件 {{.File}} 发生了变化'
cron     = '*/10 * * * *' #表示每10秒执行该lua脚本
os_arch  = ["Linux"]
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
