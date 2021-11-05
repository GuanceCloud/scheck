package system

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"
)

/*
	测试用例 多平台都兼容的方法 在这里运行
		只能在个别平台运行的 放到各自的test文件中
		比如 sysctl 只能在linux环境下 放到： public_flie_linux_test.go

*/

// nolint
func rangeDepth2Tabel(lv lua.LValue, t *testing.T) {
	if lv.Type() == lua.LTTable {
		if lt, ok := lv.(*lua.LTable); !ok {
			return
		} else {
			lt.ForEach(func(_ lua.LValue, value lua.LValue) {
				valT, ok := value.(*lua.LTable)
				if !ok {
					return
				}
				valT.ForEach(func(name lua.LValue, cron lua.LValue) {
					nameStr := lua.LVAsString(name)
					cronStr := lua.LVAsString(cron)
					t.Logf("show:  %s : %s", nameStr, cronStr)
				})
			})
		}
	}
}

func TestHostname(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "case1", args: args{l: lua.NewState()}, want: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := Hostname(test.args.l); got != test.want {
				t.Errorf("hostname() = %v, want %v", got, test.want)
			} else {
				lv := test.args.l.Get(1)
				t.Log(lua.LVAsString(lv)) // 获取hostname
			}
		})
	}
}

func TestZone(t *testing.T) {
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
			if got := Zone(test.args.l); got != test.want {
				t.Errorf("zone() = %v, want %v", got, test.want)
			} else {
				lv := test.args.l.Get(1)
				t.Log(lua.LVAsString(lv))
			}
		})
	}
}

func TestScLog(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "log test", args: args{l: lua.NewState()}, want: 0},
	}
	for _, test := range tests {
		test := test
		test.args.l.Push(lua.LString("this is test log msg..."))
		t.Run(test.name, func(t *testing.T) {
			if got := ScLog(test.args.l); got != test.want {
				t.Errorf("log() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestMounts(t *testing.T) {
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
			if got := Mounts(test.args.l); got != test.want {
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

func TestUname(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "hostInfo", args: args{l: lua.NewState()}, want: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := Uname(test.args.l); got != test.want {
				t.Errorf("uname() = %v, want %v", got, test.want)
			} else {
				lv := test.args.l.Get(1)
				if lv.Type() == lua.LTTable {
					lt, ok := lv.(*lua.LTable)
					if ok {
						lt.ForEach(func(key lua.LValue, value lua.LValue) {
							keyStr := lua.LVAsString(key)
							valStr := lua.LVAsString(value)
							t.Logf("key=%s val=%s", keyStr, valStr)
						})
						t.Log("uname  run ok ")
					}
				}
			}
		})
	}
}

func TestUptime(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "all", args: args{l: lua.NewState()}, want: 1},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			if got := Uptime(test.args.l); got != test.want {
				t.Errorf("uptime() = %v, want %v", got, test.want)
			} else {
				lv := test.args.l.Get(1)
				t.Log(lua.LVAsNumber(lv))
			}
		})
	}
}

func TestLoader(t *testing.T) {
	for i := 0; i < 100; i++ {
		go Loader(lua.NewState())
	}
	time.Sleep(time.Second)
}

func TestTicker(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "ticker-test-1s", args: args{l: lua.NewState()}, want: 0},
	}
	for _, tt := range tests {
		test := tt
		c := make(lua.LChannel, 1)
		test.args.l.Push(c)
		test.args.l.Push(lua.LNumber(1))
		t.Run(test.name, func(t *testing.T) {
			if got := Ticker(test.args.l); got != test.want {
				t.Errorf("Ticker() = %v, want %v", got, test.want)
			}
		})
		s := <-c
		t.Log(s)
		time.Sleep(time.Second * 2) // 等待测试完成
	}
}
