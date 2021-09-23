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
id         = 'users-checker'
category   = 'system'
level      = 'warn'
title      = '监视系统用户变动'
desc       = '{{.Content}}'
cron       = '*/10 * * * *'
instanceId = 'id-xxx'
os_arch    = ["Linux"]
```


3. 在清单文件同级目录下新建脚本文件 `users.lua`，编辑如下：
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

4. 当有用户被添加了，下一个 10 秒会检测到并触发 trigger 函数，从而将事件发送到文件 `/var/log/scheck/event.log` 中，添加一行数据，例：  

```
users-checker,category=system,level=warn,title=监视系统用户变动 message="新用户: xxx" 1617262230001916515
```
