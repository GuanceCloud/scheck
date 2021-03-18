package funcs

import (
	"time"

	lua "github.com/yuin/gopher-lua"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type (
	LuaFunc struct {
		Name    string
		Fn      lua.LGFunction
		Title   string
		Desc    string
		Params  []string
		Returns []string
		Test    []string
	}
)

var (
	SupportFuncs = []LuaFunc{
		{
			Name:  `file_exist`,
			Fn:    fileExist,
			Title: `file_exist(filepath)`,
			Params: []string{
				`*filepath: file location to check.`,
			},
			Returns: []string{
				`boolean: true if exists, otherwise is false`,
			},
			Desc: `
check if a file exists.`,
			Test: []string{`
file='/your/file/path'
exists=file_exist(file)
`},
		},

		{
			Name:  `read_file`,
			Fn:    readFile,
			Title: `read_file(filepath)`,
			Params: []string{
				`*filepath: the file to read.`,
			},
			Returns: []string{
				`string: file content`,
				`string: error detail if read failed`,
			},
			Desc: `reads the file content.`,
			Test: []string{`
file='/your/file/path'
content, err=read_file(file)
`},
		},

		{
			Name:  `send_data`,
			Fn:    sendLineData,
			Title: `send_data(measurement, fields[, tags[, timestamp]])`,
			Desc: `
send data with the format of influxdb line protocol.`,
			Params: []string{
				`*measurename: The name of the measurement that you want to write your data to.`,
				`*fields: The field(s) for your data point. Every data point requires at least one field in line protocol.`,
				`tags: The tag set that you want to include with your data point.`,
				`timestamp: second-precision Unix time, use current time if empty.`,
			},
			Returns: []string{
				`string: empty if success, otherwise contains the error detail.`,
			},
			Test: []string{`
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
`},
		},
	}

	moduleLogger = logger.DefaultSLogger("funcs")
)

func Init() {
	moduleLogger = logger.SLogger("funcs")
}

func sendMetric(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return nil
}
