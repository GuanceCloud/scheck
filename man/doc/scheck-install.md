
- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

# 简介

本文介绍 Scheck 的基本安装。

## 安装/更新

## *安装*： 
### Linux 平台

```Shell
sudo -- bash -c "$(curl -L https://static.dataflux.cn/scheck/install.sh)"
```

### Windows 平台
```powershell
Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.dataflux.cn/scheck/install.ps1 -destination .install.ps1; powershell .install.ps1;
```


## *更新*：  
### Linux 平台
```Shell
SC_UPGRADE=1 bash -c "$(curl -L https://static.dataflux.cn/scheck/install.sh)"
```
### Windows 平台
```powershell
$env:SC_UPGRADE;Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://static.dataflux.cn/scheck/install.ps1 -destination .install.ps1; powershell .install.ps1;
```


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
