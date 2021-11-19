English | [简体中文](readme.md)

# Security Checker

Generally, a very important work in the operation and maintenance process is to patrol a series of states such as the system,software and logs. The traditional scheme often carries out similar work by writing shell (bash) scripts by engineers, and realizes cluster management through some remote script management tools. However, this method is actually very dangerous. Due to the problem of excessive permissions in the system Patrol operation, it is often run in root mode. Once the malicious script is executed, the consequences are unimaginable.

In fact, there are two kinds of malicious scripts. One is malicious commands, such as' rm -rf ', and the other is data theft, such as leaking data to the outside through network io.

so Security Checker is hoped to provide a new secure script mode (restrict command execution,  local IO and  network IO) to ensure that all behaviors are safe and controllable, and the security checker will collect patrol events through a unified network model in the form of log. 

At the same time, security checker will provide a large number of updatable rule base scripts, including system, container, network, security and a series of patrol inspections.


## Build source code
> Because Scheck refers to the internal library, the `go mod tidy/vendor` commands cannot be used during source installation or secondary development (unrecognized import path).
> Please use gopath mode and manually put the new third-party library under ` $GOPATH/src'.

### go get 
```shell script
cd $GOPATH/src
go get -d  github.com/DataFlux-cn/scheck
```

### Dependencies
- `make`:for Makefile
- `golangci-lint`: for Makefile usage
- `packr2`: for packaging manuals
- `tree`: for Makefile manuals

###  build
Scheck is a project maintained on gitlib . so you need to migrate directories before compiling.
 
It is recommended to initialize the project environment:
```shell script
cd $GOPATH/src
mkdir -p gitlab.jiagouyun.com/cloudcare-tools/sec-checker
cp -r github.com/DataFlux-cn/scheck/. gitlab.jiagouyun.com/cloudcare-tools/sec-checker/
cp -r github.com/DataFlux-cn/scheck/vendor/. $GOPATH/src/
```

### Build local package
```
GO111MODULE=off;make local
```
> Please check the file for make command related instructions:[Makefile](Makefile)

## install&upgrade with shell
### Linux
*install*：  
```Shell
sudo -- bash -c "$(curl -L https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh)"
```

*upgrade*：  
```Shell
SC_UPGRADE=1 bash -c "$(curl -L https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.sh)"
```

After installation, it will run as a service. The service name is`scheck`. Use the service management tool to control the start / stop of the program:  

```
systemctl start/stop/restart scheck
```

or

```
service scheck start/stop/restart
```

### Windows
*install*：
```powershell
Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

*upgrade*：
```powershell
$env:SC_UPGRADE;Set-ExecutionPolicy Bypass -scope Process -Force; Import-Module bitstransfer; start-bitstransfer -source https://zhuyun-static-files-testing.oss-cn-hangzhou.aliyuncs.com/security-checker/install.ps1 -destination .install.ps1; powershell .install.ps1;
```

The default installation directory is: `/usr/local/scheck`.
The lua script directory is :`/usr/local/scheck/rules.d`,

> Be careful not to write rules:Security Checker each startup and update overwrites the file again !

## More references(detailed documentation)
- [best-practices](https://www.yuque.com/dataflux/sec_checker/best-practices)
- [more Scheck cmd](https://www.yuque.com/dataflux/sec_checker/scheck-how-to#c5609495)
- how to use lua lib and go export for lua func. [lua-lib](https://www.yuque.com/dataflux/sec_checker/lualib) , [go-openlib](https://www.yuque.com/dataflux/sec_checker/funcs)
- Now there are more than 100 rules: [rules list](https://www.yuque.com/dataflux/sec_checker/0001-user-add)
- Users can customize their own rule and create lua lib,[how to](https://www.yuque.com/dataflux/sec_checker/custom-how-to)


## config

config file at `/usr/local/scheck/scheck.conf`，profile adoption [TOML](https://toml.io/en/) :

```toml
[system]
  # ## scheck rules dir
  rule_dir = "/usr/local/scheck/rules.d"
  # custom rules dir
  custom_dir = "/usr/local/scheck/custom.rules.d"
  custom_rule_lib_dir = "/usr/local/scheck/custom.rules.d/libs"
  # lua srcipt hotupdate
  lua_HotUpdate = false
  cron = ""
  # disable scheck log output
  disable_log = false

[scoutput]
   # ##scheck rule results can be sent to local,datakit httpand aliyun sls.
   # output to datakit:default to 127..
  [scoutput.http]
    enable = true
    output = "http://127.0.0.1:9529/v1/write/security"
  [scoutput.log]
    # ##output to local storage
    enable = false
    output = "/var/log/scheck/event.log"
  # aliyun sls
  [scoutput.alisls]
    enable = false
    endpoint = ""
    access_key_id = ""
    access_key_secret = ""
    project_name = "zhuyun-scheck"
    log_store_name = "scheck"

[logging]
  # scheck log
  log = "/var/log/scheck/log"
  log_level = "info"
  rotate = 0

[cgroup]
    # cgroup enable is false
  enable = false
  cpu_max = 30.0
  cpu_min = 5.0
  mem = 0

```

> After the config file is modified, the service needs to be restarted to take effect.

## The check rule
check rules in directory: by configuration file `rule_dir` or user defined directory `custom_dir ` specify. 
Each rule corresponds to two files:

1. script file:Written in [Lua](http://www.lua.org/) language, The suffix must be `.lua`   
2. manifest file: Written in [TOML](https://toml.io/en/) , The suffix must be `.manifest`, [For details, please refer to manifest](#manifest)  

script file and manifest file  **mast have the same name**

Scheck will execute the detection script periodically. When the host event triggers some rules in the detection script, it will be reported through the function `trigger()`

For security reasons, we have closed some functions of IO and OS in Lua standard library, and made corresponding supplements and extensions in [Lua lib](https://www.yuque.com/dataflux/sec_checker/lualib) and [go openlib](https://www.yuque.com/dataflux/sec_checker/funcs)


> The rule files added by users must be placed in the `custom_dir`, otherwise they will be deleted

### manifest 
 
The manifest file is a description of the contents detected by the current rule, such as check file changes, port open and close, etc. Only the fields in the manifest file will be included in the final line protocol data. Details are as follows:
```toml
# ---------------- Required fields ---------------

# Rule ID of the event, such as: k8s-pod-001. Will be used as the indicator name of the line agreement
id = '0000-file-change'

# The classification of events is customized according to the business
category = 'system'

# The risk level of the current event is customized according to the business, such as info,warn and error
level = 'warn'

# The title of the current event, describing the detected content, such as "sensitive file change"
title = 'sensitive file change'

# Report content of current event (support template, see below for details)
desc = 'file{.filename}heve change'

# How often (crontab)
cron = '0 */5 * * *'

# OS
os_arch = ["Linux"]
# ---------------- optional field ---------------

# disabled this rule
#disabled = false

# Bring your hostname when trigger
#omit_hostname = false

# Display settings hostname
#hostname = ''

# ---------------- Custom field ---------------
# value must be string
#instanceID=''
```

### Cron
```
# ┌───────────── Second
# │ ┌───────────── Miniute
# │ │ ┌───────────── Hour
# │ │ │ ┌───────────── Day-of-Month
# │ │ │ │ ┌───────────── Month
# │ │ │ │ │                                   
# │ │ │ │ │
# │ │ │ │ │
# * * * * *
```

Example：  
`10 * * * *`   
`*/10 * * * *`  
`10 1 * * *`  
`10 */3 * * *`


### Template support

manifest file `desc` is string template,grammar: `{{.<Variable>}}`, Example
```
the file{{.FileName}}is changed,The changes are: {{.Content}}
```
  
Similarly, you can pass in a Lua "table", which will be replaced when reporting. Example
```lua
tmpl_vals={
    FileName = "/etc/passwd",
    Content = "delete user demo"
}
trigger(tmpl_vals)
```

output `desc` is：

```
file /etc/passwd changed, The changes are: delete user demo
```

## test rule
After writing the rule code, you can use 'scheck --test' to test whether the code is correct. No suffix is required   
```shell
$ scheck --test  ./rules.d/demo
```

## Create user lua-lib
cd `/usr/local/scheck/custom.rules.d/lib` and create file `common.lua` 

``` lua
module={}

function modules.Foo()
    -- func body...
end

return module
```

lua script can require common module  

``` lua
common=require("common")
common.Foo()
```

## line protocol

Security Checker output is  line protocol format. Take the rule ID as the indicator name.

### tags list

| Name       | Type   | Description                                    | Required |
| :---        |:----:  | :----                                           | :---:    |
| `title`    | string | event title                                   | true     |
| `category` | string | event  category                                      | true     |
| `level`    | string | event level, supported:`info`，`warn`，`critical` | true     |
| `host`     | string | host name                                       | false |
| `os_arch`  | string | OS arch                                    | true    |
| tags | string | manifest file custom tags                        | false    |

Several current 'category' classifications

- `network` : It mainly involves connection, port, firewall, etc
- `storage` :disk,etc
- `db` :database:(MySQL/Redis/...)
- `system` :system related
- `container` : Docker and Kubernetes

### fields list

| fields    | type   | describe     |
| ---       | :---:  | ----     |
| `message` | string | event describe |

## rule Example

### Check for changes in sensitive documents

Once the sensitive file changes, the event will be recorded in the file `/var/log/scheck/event.log'    

1. cd `/usr/local/scheck/rules.d` and create file `files.manifest` :  

```toml
id       = 'check-file'
category = 'system'
level    = 'warn'
title    = 'monitor file changes'
desc     = 'file {{.File}} have changed'
cron     = '*/10 * * * *' #Lua script is executed every 10 seconds
os_arch  = ["CentOS", "Darwin"]
```

2. create file `files.lua`:

```lua
local files={
	'/etc/passwd',
	'/etc/group'
}

local function check(file)
	local cache_key=file
	local hashval = file_hash(file)

	local old = get_cache(cache_key)
	if not old then
		set_cache(cache_key, hashval)
		return
	end

	if old ~= hashval then
		trigger({File=file})
		set_cache(cache_key, hashval)
	end
end

for i,v in ipairs(files) do
	check(v)
end
```

3. when `/etc/passwd` or `/etc/passwd` have change,The trigger function will be triggered in the next 10 seconds,This sends the data to the file `/var/log/scheck/event.log` ,Add a row of data, for example:  

```
check-file-01,category=security,level=warn,title=monitor file changes message="file /etc/passwd have changed" 1617262230001916515
```
