package net

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func rangeDepth2Tabel(lv lua.LValue, t *testing.T) {
	if lv.Type() == lua.LTTable {
		lt := lv.(*lua.LTable)
		lt.ForEach(func(_ lua.LValue, value lua.LValue) {
			valT := value.(*lua.LTable)
			valT.ForEach(func(name lua.LValue, cron lua.LValue) {
				nameStr := lua.LVAsString(name)
				cronStr := lua.LVAsString(cron)
				t.Logf("show:  %s : %s", nameStr, cronStr)
			})
		})
	}
}

func TestInterfaceAddresses(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "all os", args: args{l: lua.NewState()}, want: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := InterfaceAddresses(test.args.l); got != test.want {
				t.Errorf("mounts() = %v, want %v", got, test.want)
			} else {
				lv := test.args.l.Get(1)
				if lv.Type() == lua.LTTable {
					rangeDepth2Tabel(lv, t)
				}
			}
		})
	}
}
