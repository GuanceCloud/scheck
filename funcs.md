
[file_exist](#file_exist)  
[file_info](#file_info)  
[read_file](#read_file)  
[file_hash](#file_hash)  
[hostname](#hostname)  
[uptime](#uptime)  
[time_zone](#time_zone)  
[kernel_info](#kernel_info)  
[kernel_modules](#kernel_modules)  
[ulimit_info](#ulimit_info)  
[mounts](#mounts)  
[processes](#processes)  
[interface_addresses](#interface_addresses)  
[iptables](#iptables)  
[process_open_sockets](#process_open_sockets)  
[listening_ports](#listening_ports)  
[users](#users)  
[shadow](#shadow)  
[shell_history](#shell_history)  
[trig](#trig)  


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
| mode | number | Permission bits |
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
| cmdline | string | Complete argv |
| percent_processor_time | number | Returns elapsed time that all of the threads of this process used the processor to execute instructions in 100 nanoseconds ticks. |
| path | string | Path to executed binary |
| uid | number | Unsigned user ID |
| gid | number | Unsigned group ID |
| system_time | number | CPU time in milliseconds spent in kernel space |
| user_time | number | CPU time in milliseconds spent in user space |
| nice | number | Process nice level (-20 to 20, default 0) |
| start_time | number | Process start time in seconds since Epoch, in case of error -1 |
| threads | number | Number of threads used by process |
| state | string | Process state|

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
| exe | string | The process path or shorthand argv[0] |
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
| exe | string | The process path or shorthand argv[0] |
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

## trig

`trig([template_vals])`

trig an event and send it to target with line protocol.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| template_vals | `table` | if you use template in manifest, the values of this table will replace the template variables | false |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string` | empty if success, otherwise contains the error detail |
