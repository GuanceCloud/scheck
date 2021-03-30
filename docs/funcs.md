
### [file_exist](#file_exist)
### [file_info](#file_info)
### [read_file](#read_file)
### [file_hash](#file_hash)
### [hostname](#hostname)
### [uptime](#uptime)
### [time_zone](#time_zone)
### [kernel_info](#kernel_info)
### [send_data](#send_data)



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
| `table` | contains details of file as followings |


*file inattributes:*  
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
| `table` | see following |


*returned table:*  
| Name | Type | Description |
| --- | ---- | ---- |
| version | string | Kernel version |
| arguments | string | Kernel arguments |
| path | string | Kernel path |
| device | string | Kernel device identifier |


---

## send_data

`send_data(measurement, fields[, tags[, timestamp]])`

send data with the format of influxdb line protocol.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| measurename | `string` | The name of the measurement that you want to write your data to | true |
| fields | `table` | The field set for your data point. Every data point requires at least one field in line protocol | true |
| tags | `table` | The tag set that you want to include with your data point | false |
| timestamp | `number` | second-precision Unix time, use current time if empty | false |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string` | empty if success, otherwise contains the error detail |


*Examples:*  

``` lua
measurename='weather'
fields={
	temperature=82,
	humidity=71
}
tags={
	location='us-midwest', 
	season='summer',
	age=1,
}

err=send_data(measurename, fields) --only fields
if err ~= '' then error(err) end

err=send_data(measurename, fields, tags) --with tags
if err ~= '' then error(err) end

err=send_data(measurename, fields, os.time()) --with timestamp
if err ~= '' then error(err) end

err=send_data(measurename, fields, tags, os.time()) --with tags and timestamp
if err ~= '' then error(err) end
```