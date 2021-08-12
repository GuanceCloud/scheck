# Security Checker 连接Datakit方案
针对以上的安全问题，DataFlux 开发的 Security Checker 提供了一种新型的安全的脚本方式来保证所有的行为安全可控： 
1. Security Checker 采用 Lua 语言来编写检测脚本，定时定期地执行检测脚本。其中 Lua 运行时由 Security Checker 提供，这样就能确保脚本中只能执行安全的操作(限制命令执行，限制本地IO，限制网络IO)。  
2. Security Checker 的每个检测规则由一个脚本和清单文件组成，清单文件定义了检测结果的格式，并将以日志方式通过统一的网络模型进行上报。  

同时，Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

# 使用 DataFlux 实时收集主机的安全状态
Security Checker 支持将检测结果发送到 DataKit，所以先安装 DataKit。  

安装好 Security Checker 后，在 `/usr/local/scheck/` 目录下，编辑配置文件 `checker.conf`：

```toml
# ##(必选) 存放检测脚本的目录
rule_dir='/usr/local/scheck/rules.d'

# ##(必选) 将检测结果采集到哪里，支持本地文件或http(s)链接
# ##本地文件时需要使用前缀 file://， 例：file:///your/file/path
# ##远程server，例：http(s)://your.url
output='http://localhost:9529/v1/write/security/'

# ##(可选) 程序本身的日志配置
log='/usr/local/scheck/log'
log_level='info'
```
编辑 Security Checker 配置文件（一般位于 `/usr/local/scheck/checker.conf`），将 `output` 指向 DataKit 的时序数据接口即可：

```toml
output = 'http://localhost:9529/v1/write/security' # datakit 1.1.6(含)以上版本才支持
```
 

将编写好的检测规则，放在配置文件中 `rule_dir` 指定的目录下，Security Checker 将自动定时定期地执行。 

接下来，当检测的事件触发后(这里假设监视 Linux 下 `passwd` 文件的变动)，Security Checker 将收集的日志发送到 DataKit，Datakit 进而再发送到 DataFlux，就能在 DataFlux 平台看到对应的日志：  
![](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/security-checker_a.png)
