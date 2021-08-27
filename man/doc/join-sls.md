# Scheck 连接阿里云日志系统方案

- 版本：%s
- 发布日期：%s
- 操作系统支持：windows/amd64,windows/386,linux/arm,linux/arm64,linux/386,linux/amd64

## 前提条件

- 拥有阿里云管理RAM权限
- scheck 版本大于v1.0.1


## 获取阿里云日志服务密钥
[阿里云官方文档](https://help.aliyun.com/document_detail/29009.html?spm=a2c4g.11186623.6.1468.672b693bQhatOa)

- 创建用户

  登陆RAM访问控制-左边栏[身份管理-用户]-创建用户

  创建用户，设置为登录名称scheck，选择*Open API 调用访问*

  同时记得**保存**AccessKey ID和AccessKey Secret

- 授权

  登录RAM访问控制-左边栏[权限管理-授权]-新增授权

  将AliyunLogFullAccess授权schek账号。

## 操作步骤

### 修改配置

- Linux 主机修改/usr/local/scheck/scheck.conf ，windows 主机修改C:\\Program Files\\scheck\scheck.conf 

  ```sh
    [scoutput.alisls]
      enable = true # 启动配置
      endpoint = "cn-hangzhou.log.aliyuncs.com" #设置阿里云endpoint
      access_key_id = "LTAI5tHb2vMLV4axxxxxx"  # 上步骤获取
      access_key_secret = "FNUkk52gWsZHKauXPg24jxxxx" # 上步骤获取
      project_name = "zhuyun-scheck" # 可自定义
      log_store_name = "scheck"      # 可自定义
  ```

- 参数描述如下
  
  | 参数名称            | 示例值                         | 是否必填 | 描述                                          |
  | :------------------ | :----------------------------- | :------: | :-------------------------------------------- |
  | `enable`            | `true`                         |    是    | 配置开关                                      |
  | `endpoint`          | `cn-hangzhou.log.aliyuncs.com` |    是    | 阿里云地域                                    |
  | `access_key_id`     | `LTAI5tHb2vMLV4axxxxxx`        |    是    | 阿里云AccessKey ID（AliyunLogFullAccess权限） |
  | `access_key_secret` | `FNUkk52gWsZHKauXPg24jxxxx`    |    是    | 阿里云AccessKey Secret                        |
  | `project_name`      | `zhuyun-scheck`                |   不是   | 阿里云日志系统的项目名称                      |
  | `log_store_name`    | `scheck`                       |   不是   | 日志库名称                                    |

### 重启并检测是否生效

- 重启scheck
  
  ```sh
    $systemctl restart scheck
  ```
  
- 添加新用户测试是否是否配置成功

  ```sh
  $ useradd test
  ```

- [阿里云控制台](https://sls.console.aliyun.com/lognext/profile) 查看，zhuyun-scheck project

  ![sls.png](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/sls.png)

## Grafana 对接 阿里云日志

[官方文档](https://help.aliyun.com/document_detail/60952.html?spm=5176.21213303.J_6028563670.7.65713edaY7xSV2&scm=20140722.S_help%40%40%E6%96%87%E6%A1%A3%40%4060952.S_0.ID_60952-RL_grafana-OR_s%2Bhelpproduct-V_1-P0_0)

| 软件名称                             | 版本                                                         | 描述           |
| ------------------------------------ | ------------------------------------------------------------ | -------------- |
| grafana                              | 8.0.6                                                        | 开源展示软件   |
| aliyun-log-grafana-datasource-plugin | [2.8](https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/releases/tag/2.8?spm=a2c4g.11186623.2.13.7a703e0anzkNTh&file=2.8) | 阿里云日志插件 |
|                                      |                                                              |                |



### 安装grafana

#### 1.docker 安装grafana 

```sh
docker run \
	--name=grafana \
	--volume=~/grafana/data/:/var/lib/grafana \
	-p 3000:3000 \
	grafana/grafana:8.0.6
```

#### 2.安装和配置aliyun-log-grafana-datasource-plugin

~/grafana/data/为持久化路径

```sh
$wget -o aliyun-log-grafana-datasource-plugin-master.zip  https://github.com/aliyun/aliyun-log-grafana-datasource-plugin/releases/tag/2.8?spm=a2c4g.11186623.2.13.7a703e0anzkNTh&file=2.8
$unzip 2.8.zip
$mv aliyun-log-grafana-datasource-plugin-2.8 ~/grafana/data/plugins/aliyun-log-grafana-datasource-plugin
# 修改配置
$docker exec -i grafana  sed -i '/;allow_loading_unsigned_plugins/i\allow_loading_unsigned_plugins \= aliyun-log-service-datasource,grafana-log-service-datasource
' /etc/grafana/grafana.ini
# 重启容器
$docker restart grafana
```

#### 3.配置数据源

- 浏览器登录http://127.0.0.1:3000 用户名：admin 密码：admin
- 访问http://127.0.0.1:3000/datasources，添加数据源选择log-service-datasource，Name 设置为sc，下面继续填信息。

#### 4.导入scheck模版

- [下载模版](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/grafana/zhuyun-scheck-1629358061303.json)

- 访问http://127.0.0.1:3000/dashboard/import 上传json 模版

![](https://security-checker-prod.oss-cn-hangzhou.aliyuncs.com/img/grafana/scheck-grafana.png)

