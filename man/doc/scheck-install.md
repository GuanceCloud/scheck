
- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

# 简介

本文介绍 Scheck 的基本安装。

## 安装/更新

## *安装*： 
### Linux 平台

```Shell
sudo -- sh -c 'curl https://static.dataflux.cn/scheck/install.sh | sh'
```

### Windows 平台



## *更新*：  
### Linux 平台
- amd64 类型
```Shell
SC_UPGRADE=true;sudo -- sh -c 'curl https://static.dataflux.cn/scheck/install.sh | sh'
```
### Windows 平台



安装完成后即以服务的方式运行，服务名为`scheck`，使用服务管理工具来控制程序的启动/停止：  

```
systemctl start/stop/restart scheck
```

或

```
service scheck start/stop/restart
```


其它相关链接：

- 关于 Scheck 的基本 使用，参考 [Scheck 使用入门](scheck-how-to)
