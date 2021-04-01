
## Security Checker
一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。  
实际情况中存在两种恶意脚本，一种是恶意命令，如 `rm -rf`，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。 
因此 Security Checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 Security Checker 将以日志方式通过统一的网络模型进行巡检事件的收集。同时 Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。


## 安装/更新

*安装*：  
```Shell
bash -c "$(curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh)"
```

*更新*：  
```Shell
bash -c "$(curl https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh) --upgrade"
```

安装完成后即以服务的方式运行，可以使用 `service`，`systemctl` 等服务管理工具来控制程序的启动、停止。 

>Security Checker 目前仅支持 Linux

## 配置

进入默认安装目录 `/usr/local/security-checker`，生成配置文件(如果没有) `./security-checker -config-sample > checker.conf`，配置文件采用 [TOML](https://toml.io/en/) 格式，说明如下：

```toml
# ##(必选) 存放检测脚本的目录
rule_dir='/usr/local/security-checker/checker.conf'

# ##(必选) 将检测结果采集到哪里，支持本地文件或http(s)链接
# ##本地文件时需要使用前缀 file://， 例：file:///your/file/path
# ##远程server，例：http(s)://your.url
output=''

# ##(可选) 程序本身的日志配置
log='/usr/local/security-checker/log'
log_level='info'
```

>配置文件的修改需要重启才生效，例如：`systemctl restart security-checker` 或 `service security-checker restart`


## 检测规则
检测规则放在规则目录中：由配置文件中 `rule_dir` 指定。每项规则对应两个文件：  
1. 脚本文件：使用 [Lua](http://www.lua.org/) 语言来编写，必须以 `.lua` 为后缀名。    
2. 清单文件：使用 [TOML](https://toml.io/en/) 格式，必须以 `.manifest` 为后缀名，[详情](#清单文件)。  
脚本文件和清单文件必须同名。   

Security Checker 会定时周期性(由清单文件的 `cron` 指定)地执行检测脚本，lua脚本代码每次执行时检测相关安全事件(比如文件被改动、有新用户登录等)是否触发，如果触发了，则使用 [trig]() 函数将事件(以行协议格式)发送到由配置文件中 `output` 指定的目标中。   

Security Checker 定义了若干 lua 扩展函数，并且为确保安全，禁用了一些lua包或函数，仅支持以下 lua 内置包/函数：  
``` lua
内置基础函数，如print
table
math
string
debug
os - 其中以下函数被禁用："execute", "remove", "rename", "setenv", "setlocale"
```

>目前每次添加一项检测规则都需要重启

### 清单文件 
清单文件是对当前规则所检测内容的一个描述，比如检测文件变化，端口启停等。最终的行协议数据中只会包含清单文件中的字段。详细内容如下：    
```toml
# ##(必选)事件的规则编号，如 k8s-pod-001，将作为行协议的指标名
id=''

# ##(必选)事件的分类，根据业务自定义，如 security，network
category=''

# ##(必选)当前事件的危险等级，根据业务自定义，比如warn，error
level=''

# ##(必选)当前事件的标题，描述检测内容，如 "敏感文件改动"
title=''

# ##(必选)当前事件的内容（支持模板，详情见下）
desc=''

# ##配置事件的执行周期（使用 linux crontab 的语法规则）
cron=''
```

#### 模板支持
清单文件中 `desc` 的字符串中支持设置模板变量，语法为 `{{.<Variable>}}`，比如 `文件{{.FileName}}被修改，改动内容为: {{.Content}}` 表示 FileName 和 Diff 是模板变量，将会被替换(包括前面的点号`.`)。变量的替换发生在调用 `trig` 函数时，该函数可传入一个 `table` 变量，其中包含了模板变量的替换值，假设传入如下参数：  
```lua
tmpl_vals={
    FileName="/etc/passwd",
    Content="delete user demo"
}
trig(tmpl_vals)
```
则最终的 `desc` 值为：`文件/etc/passwd被修改，改动内容为: delete user demo`


## lua 函数
见 [函数](./funcs.md)


---

## 示例
假设需要每10秒检查文件 `/etc/passwd` 的变动，检测到发生变动后，将事件记录到文件 `/var/log/security-checker/event.log` 中，可以进行如下操作：  
1. 进入安装目录，编辑配置文件 `checker.conf` 的 `output` 字段：  
```toml
rule_dir='/usr/local/security-checker/rules.d'

output='/var/log/security-checker/event.log'

log='/usr/local/security-checker/log'
log_level='info'
```

2. 在目录`/usr/local/security-checker/rules.d`(即以上配置文件中的`rule_dir`)下新建清单文件 `demo.manifest`，编辑如下：  
```toml
id='check-file-01'
category='security'
level='warn'
title='监视文件变动'
desc='文件 {{.File}} 发生了变化'
cron='*/10 * * * *' #表示每10秒执行该lua脚本
```

3. 在清单文件同级目录下新建脚本文件 `demo.lua`，编辑如下：
```lua
function detectFileChange(file)
    --file_hash为 Security Checker 提供的函数，用于计算文件的 MD5 哈希值
	hashval = file_hash(file)

	cachekey=file

	oldval = get_cache(cachekey)
	if not oldval then
		set_cache(cachekey, hashval)
		return
	end

	if oldval ~= hashval then
		err = trig({File=file})
		if err == "" then
			set_cache(cachekey, hashval)
		else
			print("error: "..err)
		end
	end
end

detectFileChange('/etc/passwd')
```

4. 重启服务

5. 文件 `/etc/passwd'` 被改动后，下一个10秒会检测到并触发 trig 函数，从而将事件发送到文件 `/var/log/security-checker/event.log` 中，添加一行数据：  
```
check-file-01,category=security,level=warn,title=监视文件变动 message="文件 /etc/passwd 发生了变化" 1617262230001916515
```