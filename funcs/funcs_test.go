package funcs

import (
	"flag"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

var (
	flagFunName = flag.String("func", "", "lua function name to test")
)

func TestFun(t *testing.T) {

	flag.Parse()

	ls := lua.NewState()
	defer ls.Close()
	for _, f := range SupportFuncs {
		if *flagFunName != "" {
			if f.Name != *flagFunName {
				continue
			}
		}
		ls.Register(f.Name, f.Fn)
		if len(f.Test) == 0 {
			continue
		}
		moduleLogger.Debugf("Start testing '%s' ...", f.Name)
		for idx, t := range f.Test {
			moduleLogger.Debugf("Demo %d:", idx+1)
			if err := ls.DoString(t); err != nil {
				moduleLogger.Errorf("test failed: %s", err)
				return
			}
		}
		moduleLogger.Debugf("End testing '%s'", f.Name)
	}
}
