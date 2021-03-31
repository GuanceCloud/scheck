
## Security Checker
一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。  
实际情况中存在两种恶意脚本，一种是恶意命令，如 `rm -rf`，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。 
因此 Security Checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 Security Checker 将以客户端形式运行，并以日志方式通过统一的网络模型进行巡检事件的收集。同时 Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

>Security Checker 仅支持 Linux

### 安装/更新

*安装*：  
``` bash
bash -c "$(curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh)"
```

*更新*：  
``` bash
bash -c "$(curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh) --upgrade"
```

### 使用

进入安装目录 `/usr/local/security-checker`，打开主配置文件 `checker.conf`：  
``` toml
# ##(必选) 指定存放这个检测脚本的目录
rule_dir='/usr/local/security-checker/checker.conf'

# ##(必选) 配置输入以收集脚本产生的检测日志，支持本地文件或http(s)链接
# ##本地文件时需要使用前缀 file://， 例: file:///your/file/path
# ##remote:  http(s)://your.url
output=''

# ##(可选) 配置程序日志输入
log='/usr/local/security-checker/log'
log_level='info'
```

配置完成后执行 `systemctl restart security-checker` 重启即可生效。


### 脚本

检测脚本使用 lua 编写，每个检车脚本对应一个代码文件 `*.lua` 和配置文件 `*.conf`。   

**脚本配置文件**  
其中配置的字段将作为检测日志的附加字段一起发送，格式如下：    
``` toml
# ##事件的规则编号，如k8s-pod-001
rule_id=''

# ##事件的分类，如 security，network
category=''

# ##当前事件的危险等级
level=''

# ##当前事件的标题（支持模板）
title=''

# ##当前事件的内容（支持模板）
desc=''

# ##配置事件的执行周期（使用 crontab 的语法规则）
cron=''
```

**脚本代码**  
为确保安全，在执行脚本代码时禁用了一些lua包或函数，脚本中只能使用 Security Checker 提供的 lua 扩展函数，并且不支持导入第三方 lua 包。  

仅支持以下 lua 内置包/函数：  
``` lua
内置基础函数，如print
table
math
string
debug
os - 其中以下函数被禁用："execute", "remove", "rename", "setenv", "setlocale"
```

编写完脚本及其配置文件后，放入主配置文件中指定的rule_dir目录中，重启服务，Security Checker 会在指定时间或指定间隔执行检测脚本并以日志形式发送检测结果。  

### lua 函数
见 [函数](./docs/funcs.md)