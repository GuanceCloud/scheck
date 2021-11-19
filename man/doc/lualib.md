# lua 标准库和lua-lib支持清单
- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

## 说明
scheck 旨在提供一个 golang 所使用的功能完善、集成简单的 lua 沙盒环境，以适应用户不同的需求和环境。
所以我们引用了一款优秀的开源lua VM：[gopher-lua](https://github.com/yuin/gopher-lua) ,其所使用的 lua 沙盒环境为 lua 5.1 版本。

出于安全考虑，我们关闭了lua标准库中IO和OS的部分功能，并在lua-lib和go-openlib中做了相应的补充和扩展。

本文重点介绍 在lua脚本开发过程中各种库的使用及注意的事项。

lua脚本开发中可从三个地方引用的库 分别为：
- 标准库:支持大部分库和函数。
- lua-lib：封装了一些lua常用的函数lib供开发者调用。
- go-openlib：使用golang开发并提供了`file`,`system`,`net`,`utils`等。可查看：[func清单列表](funcs)


## 标准库
关闭的函数列表：
- 所有的IO模块功能全部禁止调用。
- OS库中关闭的函数：`exit`, `execute`, `remove`, `rename`, `setenv`, `setlocale`
- OS库中可使用的函数 : `clock`,`difftime`,`date`,`getenv`,`time`,`tmpname`

## scheck中的lua-lib使用
文件位置：在安装目录下的`ruls.d/libs`下。具体函数的方法可查看相应的文件

方法列表
- common
    - watcher(path,enum,func)
- directorymonitoy
    - change(dir)     
    - add(dir)
    - del(dir)
    - priv_change(dir)
- filemonitor
    - check(file)
    - exist(file)
    - priv_change(file)
    - priv_fixedchange(file,mode)
    - priv_limit(file,mode_bits)
    - priv_root_ownership(file)
    - priv_ownership(file,user)
- kernelmonito
    - module_isinstall(module_name, isInstalled)     
- mountflagmonitor
    - check(mountpath, mountflag, value)
- rpmmonitor
    - check(pck_name, switch)
- sysctlmonitor
    - check(param, switch)
    - check_many(params, switch)


### 示例：

使用lua-lib中的方法
``` lua
local filemonitor = require("filemonitor")
local function check(file)
    if filemonitor.exist(file) then
        filemonitor.check(file) -- 监控文件 发生变化后 上报消息
    end
end
check('/etc/shadow')
```


