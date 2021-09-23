# Scheck 配置

- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

## 配置描述
  进入默认安装目录 `/usr/local/scheck`，打开配置文件 `scheck.conf`，配置文件采用 [TOML](https://toml.io/en/) 格式，说明如下：

```toml
[system]
  # ##(必选) 系统存放检测脚本的目录
  rule_dir = "/usr/local/scheck/rules.d"
  # ##客户自定义目录
  custom_dir = "/usr/local/scheck/custom.rules.d"
  #热更新
  lua_HotUpdate = false
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
### system 模块

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
```

| 参数名称      |  类型  | 描述                          |
| :------------ | :----: | ----------------------------- |
| rule_dir      | string | 系统存放检测脚本的目录        |
| custom_dir    | string | 客户自定义目录                |
| lua_HotUpdate |  bool  | 热更新，支持没10秒加载lua脚本 |
| cron          | string | 强制所以定时时间              |
| disable_log   |  bool  | 是否禁用日志                  |

### scoutput 模块

```toml
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
```

| 参数名称          |  类型  | 描述                   |
| ----------------- | :----: | ---------------------- |
| scoutput.http     |        | http 输出模块          |
| enable            |  bool  | 是否启用               |
| output            | string | datakit api 地址       |
| scoutput.log      |        |                        |
| enable            |  bool  | 是否启用               |
| output            | string | 文件路径               |
| scoutput.alisls   |        |                        |
| enable            |  bool  | 是否启用               |
| endpoint          | string | 阿里云地域             |
| access_key_id     | string | 阿里云 AccessKey ID    |
| access_key_secret | string | 阿里云AccessKey Secret |
| project_name      | string | 项目名称               |
| log_store_name    | string | 日志库名称             |
|                   |        |                        |

### logging 模块

```toml
[logging]
  # ##(可选) 程序运行过程中产生的日志存储位置
  log = "/var/log/scheck/log"
  log_level = "info"
  rotate = 0
```

| 参数名称  |  类型  | 描述                                 |
| --------- | :----: | ------------------------------------ |
| log       | string | scheck系统日志路径                   |
| log_level | string | scheck日志级别                       |
| rotate    |  int   | 0 为默认，日志分片大小 单位M 默认30M |

### cgroup 模块

```toml
[cgroup]
    # 可选 默认关闭 可控制cpu和mem
  enable = false
  cpu_max = 30.0
  cpu_min = 5.0
  mem = 0
```

| 参数名称 | 类型  | 描述        |
| -------- | ----- | ----------- |
| enable   | bool  | 是否开启    |
| cpu_max  | float | cpu最大限制 |
| mem      | float | cpu最小限制 |

