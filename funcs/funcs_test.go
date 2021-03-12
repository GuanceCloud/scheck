package funcs

import (
	"fmt"
	"log"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestFun(t *testing.T) {

	ls := lua.NewState()
	defer ls.Close()
	for _, f := range SupportFuncs {
		ls.Register(f.Name, f.Fn)
		if len(f.Test) == 0 {
			continue
		}
		log.Printf("Start testing '%s' ...", f.Name)
		for idx, t := range f.Test {
			log.Printf("Demo %d", idx+1)
			ls.DoString(t)
			fmt.Println()
		}
		log.Printf("End testing '%s'", f.Name)
	}
}
