
### [file_exist](#file_exist)
### [read_file](#read_file)
### [send_data](#send_data)



## file_exist

`file_exist(filepath)`

check if a file exists.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | file location to check | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `boolean` | `true` if exists, otherwise is `false` |

*Examples:*  

``` lua
file = '/your/file/path'
exists = file_exist(file)
print(exists)
```

---

## read_file

`read_file(filepath)`

reads the file content.

*Parameters:*  

| Name | Type | Description | Required |
| --- | ---- | ---- | ---- |
| filepath | `string` | the file path to read | true |


*Return value(s):*  

| Type | Description |
| --- | ---- |
| `string` | file content |
| `string` | error detail if read failed |


*Examples:*  

``` lua
file='/your/file/path'
content, err = read_file(file)
if err ~= '' then
    print("error: "..err)
else
    print(content)
end
```

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