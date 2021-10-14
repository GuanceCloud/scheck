# Scheck 最佳实践

- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

# 简介

    一般在运维过程中有非常重要的工作就是对系统，软件，包括日志等一系列的状态进行巡检，传统方案往往是通过工程师编写shell（bash）脚本进行类似的工作，通过一些远程的脚本管理工具实现集群的管理。但这种方法实际上非常危险，由于系统巡检操作存在权限过高的问题，往往使用root方式运行，一旦恶意脚本执行，后果不堪设想。实际情况中存在两种恶意脚本，一种是恶意命令，如rm -rf，另外一种是进行数据偷窃，如将数据通过网络 IO 泄露给外部。 因此 Security Checker 希望提供一种新型的安全的脚本方式（限制命令执行，限制本地IO，限制网络IO）来保证所有的行为安全可控，并且 Security Checker 将以日志方式通过统一的网络模型进行巡检事件的收集。同时 Security Checker 将提供海量的可更新的规则库脚本，包括系统，容器，网络，安全等一系列的巡检。

> scheck 为Security Checker 简称
>
> scheck 只推送安全巡检事件，没有恢复通知



# 前置条件


| 服务名称 | 版本                                                         | 是否必须安装 | 用途            |
| -------- | ------------------------------------------------------------ | ------------ | --------------- |
| Datakit  | 1.1.6 以上 [安装方法](https://www.yuque.com/dataflux/datakit/datakit-install) | 必须         | 接受scheck 信号 |
| DataFlux | [DataFlux SaaS](https://dataflux.cn) 或其他私有部署版本      | 必须         | 查看安全巡检    |




# 配置

### 1 安装 Scheck

```sh
sudo -- bash -c "$(curl -L https://static.dataflux.cn/security-checker/install.sh)"
```



### 2 查看安装状态以及datakit运行状态
- 查看scheck 状态
```sh
$systemctl status scheck
● scheck.service - security checker with lua script
   Loaded: loaded (/usr/lib/systemd/system/scheck.service; enabled; vendor preset: disabled)
   Active: active (running) since 六 2021-07-03 00:13:15 CST; 2 days ago
 Main PID: 15337 (scheck)
    Tasks: 10
   Memory: 12.4M
   CGroup: /system.slice/scheck.service
           └─15337 /usr/local/scheck/scheck -config /usr/local/scheck/scheck.conf
           
```
- 查看datakit 状态
```shell
$ systemctl status datakit
● datakit.service - Collects data and upload it to DataFlux.
   Loaded: loaded (/etc/systemd/system/datakit.service; enabled; vendor preset: disabled)
   Active: active (running) since 六 2021-07-03 01:07:44 CST; 2 days ago
 Main PID: 27371 (datakit)
    Tasks: 9
   Memory: 29.6M
   CGroup: /system.slice/datakit.service
           └─27371 /usr/local/datakit/datakit
```


### 3 登录DataFlux控制台查看安全巡检记录（[Saas平台](https://dataflux.cn)）

- 选择左侧栏-安全巡检 查看巡检内容	![1](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/bestpractices/2.png)





# 相关命令
> Security Checker cmd
- 查看帮助
```sh
$scheck -h
Usage of scheck:
  -check-md5
    	md5 checksum
  -config string
    	configuration file to load
  -config-sample
    	show config sample
  -funcs
    	show all supported lua-extend functions
  -test string
    	the name of a rule, without file extension
  -testc int
    	test rule count
  -version
    	show version
  -doc 
        Generate doc document from manifest file
  -tpl
        Generate doc document from template file
  -dir
        配合`-doc` `-tpl`使用，将文件输出到指定的目录上
  -luastatus
        展示所有的lua运行状态，并输出到当前的目录下，文件的格式Markdown合适。
  -sort
        配合`-luastatus`使用，排序的参数有：名称：name,运行时长：time,运行次数：count,默认使用count
     ./scheck -luastatus -sort=time
  -check
        预编译一次用户目录下所有的lua文件，检查代码的语法性错误。
  -box
        展示所有加载到二进制中的文件列表
```


- 启停命令 
```sh
systemctl start/stop/restart/status scheck 
## 或者 
service scheck start/stop/restart/status 
``` 

