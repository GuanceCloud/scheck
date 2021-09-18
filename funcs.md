
1. [ls](#ls)   
2. [file_exist](#file_exist)  
3. [file_info](#file_info)  
4. [read_file](#read_file)  
5. [file_hash](#file_hash)  
6. [hostname](#hostname)  
7. [uptime](#uptime)  
8. [time_zone](#time_zone)  
9. [kernel_info](#kernel_info)  
10. [kernel_modules](#kernel_modules)  
11. [ulimit_info](#ulimit_info)  
12. [mounts](#mounts)  
13. [interface_addresses](#interface_addresses)  
14. [iptables](#iptables)  
15. [processes](#processes)  
16. [process_open_sockets](#process_open_sockets)  
17. [process_open_files](#process_open_files)  
18. [listening_ports](#listening_ports)  
19. [users](#users)  
20. [logged_in_users](#logged_in_users)  
21. [last](#last)  
22. [lastb](#lastb)  
23. [shadow](#shadow)  
24. [shell_history](#shell_history)  
25. [trigger](#trigger)  
26. [get_cache](#get_cache)   
27. [set_cache](#set_cache)  
28. [json_encode](#json_encode)   
29. [json_decode](#json_decode)  
30. [crontab](#crontab)  
31. [uname](#uname)  
32. [sysctl](#sysctl)  
33. [rpm_list](#rpm_list)  
34. [rpm_query](#rpm_query)  
35. [grep](#grep)  
36. [get_global_cache](#get_global_cache)  
37. [set_global_cache](#set_global_cache)  
38. [mysql_weak_psw](#mysql_weak_psw)  
39. [mysql_ports_list](#mysql_ports_list)  
40. [sc_sleep](#sc_sleep)
41. [sc_path_watch](#sc_path_watch)
42. [sc_ticker](#sc_ticker)

## ls

`ls(dir[, rescue])`

list files in specified directory.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| dir | `string` | path of dir  | true |
| rescue | `boolean` | if recursively traverse the dir, default is false  | false |

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| path | string | file's full path |
| filename | string | name of file |
| size | number | Size of file in bytes |
| block_size | number | Block size of filesystem |
| mode | string | Permission bits |
| uid | number | Owning user ID |
| gid | number | Owning group ID |
| device | number | Device ID (optional) |
| inode | number | Filesystem inode number |
| hard_links | number | Number of hard links |
| ctime | number | Last status change time |
| mtime | number | Last modification time |
| atime | number | Last access time |

---

## file_exist

`file_exist(filepath)`

check if a file exists.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | absolute file path  | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `boolean` | `true` if exists, otherwise is `false` |

*Example:*  

``` lua
file = '/your/file/path'
exists = file_exist(file)
print(exists)
```

---

## file_info

`file_info(filepath)`

read file attributes and metadata.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | absolute file path  | true |


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table` | contains details of file as below |


| Name | Type | Description |
| --- | ---- | ---- |
| size | number | Size of file in bytes |
| block_size | number | Block size of filesystem |
| mode | string | Permission string  |
| perm | string | Permission bits |
| uid | number | Owning user ID |
| gid | number | Owning group ID |
| device | number | Device ID (optional) |
| inode | number | Filesystem inode number |
| hard_links | number | Number of hard links |
| ctime | number | Last status change time |
| mtime | number | Last modification time |
| atime | number | Last access time |

*Example:*  

``` lua
file = '/your/file/path'
info = file_info(file)
```

---


## read_file

`read_file(filepath)`

reads the file content.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | absolute file path| true |


*Return value(s):*   

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `string` | file content |


*Examples:*  

``` lua
file='/your/file/path'
content = read_file(file)
```

---


## file_hash

`file_hash(filepath)`

calculate the md5 sum of file content.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | absolute file path| true |


*Return value(s):*   

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `string` | md5 of file content |


*Examples:*  

``` lua
file='/your/file/path'
content = file_hash(file)
```

---

## hostname

`hostname()`

get current hostname.


*Return value(s):*   

it issues an error when fail to get.

| Type | Description |
| --- | ---- |
| `string` | hostname |


---

## uptime

`uptime()`

time passed since last boot.


*Return value(s):*   

| Type | Description |
| --- | ---- |
| `number` | Total uptime seconds |

---

## time_zone

`time_zone()`

current timezone in the system


*Return value(s):*   

| Type | Description |
| --- | ---- |
| `string` | current timezone in the system |

---


## kernel_info

`kernel_info()`

linux kernel modules both loaded and within the load search path


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table` | details as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| version | string | Kernel version |
| arguments | string | Kernel arguments |
| path | string | Kernel path |
| device | string | Kernel device identifier |

---

## kernel_modules

`kernel_modules()`

linux kernel modules both loaded and within the load search path


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a module describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| name | string | Module name|
| size | string | Size of module content |
| used_by | string | Module reverse dependencies |
| status | string | Kernel module status |
| address | string | Kernel module address |

---

## ulimit_info

`ulimit_info()`

System resource usage limits.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table` | see below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| type | string | System resource to be limited |
| soft_limit | string | Current limit value |
| hard_limit | string | Maximum limit value |

---

## mounts

`mounts()`

System mounted devices and filesystems (not process specific)


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a mounted device describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| device | string | Mounted device |
| path | string | Mounted device path |
| type | string | Mounted device type |
| flags | string | Mounted device flags |

---

## processes

`processes()`

All running processes on the host system

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a process describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| name | string | The process path or shorthand argv[0] |
| path | string | Path to executed binary |
| cmdline | string | Complete argv |
| state | string | Process state |
| cwd | string | Process current working directory |
| root | string | Process virtual root directory |
| uid | number | Unsigned user ID |
| gid | number | Unsigned group ID |
| euid | number | Unsigned effective user ID |
| egid | number | Unsigned effective group ID |
| suid | number | Unsigned saved user ID |
| sgid | number | Unsigned saved group ID |
| on_disk | number | The process path exists yes=1, no=0, unknown=-1 |
| resident_size | number | Bytes of private memory used by process |
| total_size | number | Total virtual memory size |
| system_time | number | CPU time in milliseconds spent in kernel space |
| user_time | number | CPU time in milliseconds spent in user space |
| disk_bytes_read | number | Bytes read from disk |
| disk_bytes_written | number | Bytes written to disk |
| start_time | number | Process start time in seconds since Epoch, in case of error -1 |
| parent | number | Process parent's PID |
| pgroup | number | Process group |
| threads | number | Number of threads used by process |
| nice | number | Process nice level (-20 to 20, default 0) |

---

## interface_addresses

`interface_addresses()`

Network interfaces and relevant metadata.


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a net interface describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| interface | string | Interface name |
| ip4 | string | ip4 addr |
| ip6 | string | ip6 addr |
| mtu | number | MTU |
| mac | string | MAC address |

---

## iptables

`iptables()`

Linux IP packet filtering and NAT tool

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a filtering describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| filter_name | string | Packet matching filter table name. |

---

## process_open_sockets

`process_open_sockets()`

Processes which have open network sockets on the system

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a processe which have open network describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| process_name | string | The process name |
| cmdline | string | Complete argv |
| fd | number | Socket file descriptor number |
| socket | number | Socket handle or inode number |
| family | string | Network protocol (AF_INET, AF_INET6, AF_UNIX) |
| protocol | string | Transport protocol (tcp, udp...) |
| local_address | string | Socket local address |
| remote_address | string | Socket remote address |
| local_port | number | Socket local port |
| remote_port | number | Socket remote port |
| path | string | For UNIX sockets (family=AF_UNIX), the domain path |
| net_namespace | number | The inode number of the network namespace |
| state | string | TCP socket state |

---

## process_open_files

`process_open_files()`

File descriptors for each process.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item corresponding to a processe which have open file, describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| fd | number | Process-specific file descriptor numbe |
| path | string | Filesystem path of descriptor |

---

## listening_ports

`listening_ports()`

Processes with listening (bound) network sockets/ports


*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| process_name | string | The process name |
| cmdline | string | Complete argv |
| socket | number | Socket handle or inode number |
| fd | number | Socket file descriptor number |
| address | string | Specific address for bind |
| port | number | Transport layer port |
| family | string | Network protocol (AF_INET, AF_INET6, AF_UNIX) |
| protocol | string | Transport protocol (tcp, udp...) |
| path | string | For UNIX sockets (family=AF_UNIX), the domain path |

---

## users

`users()`

Local user accounts (including domain accounts that have logged on locally (Windows))

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| username | string | Username |
| uid | number | User ID |
| gid | number | Group ID (unsigned) |
| directory | string | User's home directory |
| shell | string | User's configured default shell |


---

## logged_in_users

`logged_in_users()`

Users with an active shell on the system.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| username | string | Username |
| type | number | Login type, see utmp.h |
| tty | number | Device name |
| host | string | Remote hostname |
| time | string | Time entry was made, unix timestamp in seconds |


---

## last

`last()`

System logins and logouts.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| username | string | Username |
| type | number | Login type, see utmp.h |
| tty | number | Device name |
| host | string | Remote hostname |
| time | string | Time entry was made, unix timestamp in seconds |


---

## lastb

`lastb()`

Failed logins.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| username | string | Username |
| type | number | Login type, see utmp.h |
| tty | number | Device name |
| host | string | Remote hostname |
| time | string | Time entry was made, unix timestamp in seconds |

---

## shadow

`shadow()`

Local system users encrypted passwords and related information. Please note, that you usually need superuser rights to access `/etc/shadow`.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| username | string | Username |
| password_status | string | Password status |
| last_change | number | Date of last password change (starting from UNIX epoch date) |
| expire | number | Number of days since UNIX epoch date until account is disabled |

---

## shell_history

`shell_history()`

A line-delimited (command) table of per-user .*_history data.

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| uid | number | Shell history owner |
| history_file | string | Path to the .*_history for this user |
| command | string | Unparsed date/line/command history line |
| time | number | Entry timestamp. It could be absent, default value is 0. |

---

## trigger

`trigger([template_vals])`

trigger an event and send it to target with line protocol.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| template_vals | `table` | if you use template in manifest, the values of this table will replace the template variables | false |


*Return value(s):*  

it issues an error when failed.

---

## get_cache

`get_cache(key)`

get value for cache key.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| key | `string` | name of cache key | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string`/`boolean`/`number` | cache value |

---

## set_cache

`set_cache(key, value)`

set a cache key-value pair.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| key | `string` | name of cache key | true |
| value | `string`/`boolean`/`number` |  cache value by key | true |


*Return value(s):*  


| Type | Description |
| --- | ---- |
| `string` | error detail if failed |

---

## json_encode

`json_encode(object)`

encode a lua table to json string

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| object | `table` | a lua table  | true |


*Return value(s):*  

it issues an error if fail to encode

| Type | Description |
| --- | ---- |
| `string` | json string |

---


## json_decode

`json_decode(str)`

decode a json string to lua table

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| str | `string` | a json string  | true |


*Return value(s):*  

it issues an error if fail to encode

| Type | Description |
| --- | ---- |
| `table` | a lua table |

---

## crontab

`crontab()`

Line parsed values from system and user cron/tab.

*Return value(s):*  

it issues an error when failed.

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| minute | string | The exact minute for the job |
| hour | string | The hour of the day for the job |
| day_of_monthmand | string | The day of the month for the job |
| month | string | The month of the year for the job |
| day_of_week | string | The day of the week for the job |
| command | string | Raw command string |
| path | string | File parsed |

---

## uname

`uname()`

the operating system name and version

*Return value(s):*  

it issues an error when failed.

| Type | Description |
| --- | ---- |
| `table` | see below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| platform | string | OS Platform or ID, eg., centos |
| platform_version | string | OS Platform version, eg., 7.7.1908 |
| family | string | OS Platform family, eg., rhel |
| os | string | os name, eg., Linux |
| arch | string | OS Architecture, eg., x86_64 |
| kernel_version | string | os kernel version |

---

## sysctl

`sysctl([key])`

the operating system sysctl info

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| key | `string` | specify a key to get back, otherwise return all key-values | false |

*Return value(s):*  

it issues an error when failed.

| Type | Description |
| --- | ---- |
| `table` | same as run linux command 'sysctl -a' |

---

## rpm_list

`rpm_list()`

list all current rpm packages

*Return value(s):*  

it issues an error when failed.

| Type | Description |
| --- | ---- |
| `string` | same as run linux command 'rpm -qa' |

---

## rpm_query

`rpm_query(pkg)`

check a package is installed

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| package | `string` | the package name, eg. yum | false |

*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string` | package's fullname, or empty if not found |

---

## grep

`grep(option, pattern, file)`

run grep command

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| option | `string` | option(s) for grep | false |
| pattern | `string` | pattern for grep | true |
| file | `string` | file to search by grep | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string` | result of grep, empty if not found |
| `string` | error info if failed |

---

## mysql_weak_psw

`mysql_weak_psw(host, port [,username])`

check mysql weak password

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| host | `string` | url of mysql | true |
| port | `string` | mysql port| true |
| username | `string` | mysql username, default is 'root' | false |


*Return value(s):*  

it issues an error when failed.

| Type | Description |
| --- | ---- |
| `boolean` | true means found some weak password |
| `string` | the weak password if found |


---

## get_global_cache

`get_global_cache(key)`

get value for global cache key.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| key | `string` | name of cache key | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string`/`boolean`/`number` | cache value |

---

## set_global_cache

`set_global_cache(key, value)`

set a key-value to global cache.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| key | `string` | name of cache key | true |
| value | `string`/`boolean`/`number` |  cache value by key | true |


*Return value(s):*  


| Type | Description |
| --- | ---- |
| `string` | error detail if failed |


## mysql_ports_list

`mysql_ports_list()`

list the host MySQL ports

*Return value(s):*  

it issues an error when fail to read.

| Type | Description |
| --- | ---- |
| `table`() | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| pid | number | Process (or thread) ID |
| cmdline | string | Complete argv |
| port | number | Transport layer port |
| protocolversion | string | Mysql protocol version |
| statusflags | string | Socket file descriptor number |
| authpluginname | string | Auth plugin name |
| s️erverversion | string | Mysql S️erver Version |
| state | string | Process state |

*output example:*
```
port	3307
protocolversion	10
statusflags	2
authpluginname	mysql_native_password
s️erverversion	5.7.34
state	LISTEN
cmdline	/usr/bin/docker-proxy -proto tcp -host-ip 0.0.0.0 -host-port 3307 -container-ip 172.18.0.4 -container-port 3306
pid	7062
``` 

## sc_sleep

`sc_sleep(time)`

Thread sleeps for a certain number of seconds

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| time | `int` | second sleep time | true |


*Return value(s):*  

it issues an error when failed.


## sc_path_watch

`sc_path_watch(dir,chan)`

watch whether the file or directory has changed.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| dir | `string` | dir or filename | true |
| chan | `lua.LChannel` | lua channel | true |


*Return value(s):*  

No return value and never stop.
If the dir or file changes, it will be notified through to lua.LChannel.

*the lua.LChannel:* 

| Type | Description |
| --- | ---- |
| `table`(array) | each item describe as below |

 
| Name | Type | Description |
| --- | ---- | ---- |
| path | string | file's full path |
| status | int | file status |
*file status:*

| status | name | Description |
| --- | ---- | ---- |
| 1 | CREATE | create file in dir |
| 2 | WRITE | write file  |
| 4 | REMOVE | when file is del |
| 8 | RENAME | file rename |
| 16 | CHMOD | chmod is change |
  

## sc_ticker

 `sc_ticker(channel,time)`
 
 Send signals to the lua.LChannel regularly
 
 *Parameters:*  
 
 | Name | Type | Description | Required |
 | --- | ---- | ---- | ---- |
 | chan | `lua.LChannel` | lua channel | true |
 | time | `int` | second time | true |
 
 *Return value(s):*  
 
 No return value
 sent Lua.LString to lua.LChannel.
 
