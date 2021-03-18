package funcs

import (
	"time"

	lua "github.com/yuin/gopher-lua"

	"gitlab.jiagouyun.com/cloudcare-tools/cliutils/logger"
)

type (
	LuaFunc struct {
		Name string
		Fn   lua.LGFunction
	}
)

var (
	SupportFuncs = []LuaFunc{
		{`file_exist`, fileExist},
		{`file_info`, file_info},
		{`read_file`, readFile},
		{`send_data`, sendLineData},
	}

	moduleLogger = logger.DefaultSLogger("funcs")
)

func sendMetric(measurement string, tags map[string]string, fields map[string]interface{}, t ...time.Time) error {
	return nil
}
