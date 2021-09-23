# Scheck 连接Datakit方案

- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64


针对以上的安全问题，DataFlux 开发的 Scheck 提供了一种新型的安全的脚本方式来保证所有的行为安全可控： 
1. Scheck 采用 Lua 语言来编写检测脚本，定时定期地执行检测脚本。其中 Lua 运行时由 Scheck 提供，这样就能确保脚本中只能执行安全的操作(限制命令执行，限制本地IO，限制网络IO)。  
2. Scheck 的每个检测规则由一个脚本和清单文件组成，清单文件定义了检测结果的格式，并将以日志方式通过统一的网络模型进行上报。  

同时，Scheck 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

# 前提条件

| 服务名称 | 版本                                                         | 是否必须安装 | 用途            |
| -------- | ------------------------------------------------------------ | ------------ | --------------- |
| Datakit  | 1.1.6 以上 [安装方法](https://www.yuque.com/dataflux/datakit/datakit-install) | 必须         | 接受scheck 信号 |

# 使用 DataFlux 实时收集主机的安全状态(默认开启)
Scheck 支持将检测结果发送到 DataKit，所以先安装 DataKit。  

安装好 Scheck 后，在 `/usr/local/scheck/` 目录下，编辑配置文件 `scheck.conf`：

```toml
...
[scoutput]
   # ##安全巡检过程中产生消息 可发送到本地、http、阿里云sls。
   # ##远程server，例：http(s)://your.url
  [scoutput.http]
    enable = true
    output = "http://127.0.0.1:9529/v1/write/security"
...

```
编辑 Scheck 配置文件（一般位于 `/usr/local/scheck/scheck.conf`），将 `output` 指向 DataKit 并将`enable` 设置为`true` 的时序数据接口即可：

```toml
output = 'http://localhost:9529/v1/write/security' # datakit 1.1.6(含)以上版本才支持
```

将编写好的检测规则，放在配置文件中 `rule_dir` 指定的目录下，Scheck 将自动定时定期地执行。 

接下来，当检测的事件触发后(这里假设监视 Linux 下 `passwd` 文件的变动)，Scheck 将收集的日志发送到 DataKit，Datakit 进而再发送到 DataFlux，就能在 DataFlux 平台看到对应的日志：  
![](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/security-checker_a.png)
