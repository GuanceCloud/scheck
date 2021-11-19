package lua

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/container"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/file"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/net"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/system"
	"gitlab.jiagouyun.com/cloudcare-tools/sec-checker/funcs/utils"
)

/*
原来的形式是初始化lua.state时 将所有的函数注册进去，造成了函数的冗余，
随着函数的增多 这样的冗余的情况会越来越来越严重，所以就做了分包及模块化
-----------
把go的函数按包的形式分成多个module，在lua脚本使用module的时候 可以import导入，
具体代码介绍：
	lua：
		local m = require("mymodule")
    	m.myfunc()
    	print(m.name)
	go：
		初始化lua.state时候，需要加载openlib（gopher-lua自带）
		然后加载funcs下的lib，也是用openlib的形式

*/

type GoOpenLibs struct {
	// 模块：需要按照模块注册 引用是必须使用require()
	modules map[string]lua.LGFunction

	// 全局方法 可在lua直接调用
	globalFn map[string]lua.LGFunction
}

var goOpenLibsMux = &GoOpenLibs{modules: make(map[string]lua.LGFunction), globalFn: make(map[string]lua.LGFunction)}

func InitModules() {
	goOpenLibsMux.modules["file"] = file.Loader
	goOpenLibsMux.modules["system"] = system.Loader
	goOpenLibsMux.modules["net"] = net.Loader
	goOpenLibsMux.modules["container"] = container.Loader

	goOpenLibsMux.modules["utils"] = utils.UtilsLoader
	goOpenLibsMux.modules["json"] = utils.JSONLoader
	goOpenLibsMux.modules["cache"] = utils.CacheLoader
	goOpenLibsMux.modules["mysql"] = utils.MysqlLoader
	utils.InitLog()
}

func HandleFunc(pattern string, f func(state *lua.LState) int) {
	goOpenLibsMux.globalFn[pattern] = f
}

func LoadModule(l *lua.LState) {
	for name, module := range goOpenLibsMux.modules {
		l.PreloadModule(name, module)
	}
	for name, fn := range goOpenLibsMux.globalFn {
		l.Register(name, fn)
	}
}

func ShowModule() {
	for moduleName := range goOpenLibsMux.modules {
		state := lua.NewState()
		goOpenLibsMux.modules[moduleName](state)
		lv := state.Get(1)
		if lv.Type() == lua.LTTable {
			lt, ok := lv.(*lua.LTable)
			if ok {
				fmt.Printf("- %s :\n", moduleName)
				lt.ForEach(func(lname lua.LValue, lfn lua.LValue) {
					name := lua.LVAsString(lname)
					fmt.Printf("		- %s \n", name)
				})
			}
		}
	}
	for funcName := range goOpenLibsMux.globalFn {
		fmt.Println("global func is :", funcName)
	}
}
