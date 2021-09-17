package system

import (
	"runtime"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

/*
	测试用例 多平台都兼容的方法 在这里运行
		只能在个别平台运行的 放到各自的test文件中
		比如 sysctl 只能在linux环境下 放到： public_flie_linux_test.go

*/

const (
	winOS   = "windows"
	linuxOS = "linux"
)

func Test_provider_fileExist(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: winOS, args: args{l: lua.NewState()}, want: 1},
		{name: linuxOS, args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		if tt.name == winOS {
			tt.args.l.Push(lua.LString("C:\\users"))
		}
		if tt.name == linuxOS {
			tt.args.l.Push(lua.LString("/usr/local/bin"))
		}
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.fileExist(tt.args.l); got != tt.want {
				t.Errorf("fileExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_fileInfo(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: winOS, args: args{l: lua.NewState()}, want: 1},
		{name: linuxOS, args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		if runtime.GOOS == winOS && tt.name == winOS {
			tt.args.l.Push(lua.LString("C:\\Windows\\System32\\drivers\\etc\\hosts"))
			t.Run(tt.name, func(t *testing.T) {
				p := &provider{}
				if got := p.fileInfo(tt.args.l); got != tt.want {
					t.Errorf("fileInfo() = %v, want %v", got, tt.want)
				} else {
					lv := tt.args.l.Get(1)
					if lv.Type() == lua.LTTable {
						lt := lv.(*lua.LTable)
						lt.ForEach(func(key lua.LValue, value lua.LValue) {
							keyStr := lua.LVAsString(key)
							valStr := lua.LVAsString(value)
							t.Logf("key=%s val=%s", keyStr, valStr)

						})
					}
				}
			})
		}

		if runtime.GOOS == linuxOS && tt.name == linuxOS {
			tt.args.l.Push(lua.LString("/etc/hosts"))
			t.Run(tt.name, func(t *testing.T) {
				p := &provider{}
				if got := p.fileInfo(tt.args.l); got != tt.want {
					t.Errorf("fileInfo() = %v, want %v", got, tt.want)
				} else {
					lv := tt.args.l.Get(1)
					if lv.Type() == lua.LTTable {
						lt := lv.(*lua.LTable)
						lt.ForEach(func(key lua.LValue, value lua.LValue) {
							keyStr := lua.LVAsString(key)
							valStr := lua.LVAsString(value)
							t.Logf("key=%s val=%s", keyStr, valStr)

						})
					}
				}
			})
		}
		// windows and linux . darwin is not used.
	}
}

func Test_provider_hostname(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: winOS, args: args{l: lua.NewState()}, want: 1},
		{name: linuxOS, args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.hostname(tt.args.l); got != tt.want {
				t.Errorf("fileExist() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				t.Log(lua.LVAsString(lv)) // 获取hostname

			}
		})
	}
}

func Test_provider_zone(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.zone(tt.args.l); got != tt.want {
				t.Errorf("zone() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				t.Log(lua.LVAsString(lv))
			}
		})
	}
}

func Test_provider_log(t *testing.T) {
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
	for _, tt := range tests {
		tt.args.l.Push(lua.LString("this is test log msg..."))
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.log(tt.args.l); got != tt.want {
				t.Errorf("log() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				msg := lua.LVAsString(lv)
				t.Log(msg)
			}
		})
	}
}

func Test_provider_mounts(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.mounts(tt.args.l); got != tt.want {
				t.Errorf("mounts() = %v, want %v", got, tt.want)
			} else {
				// 返回的是二维table 把value 转成table再进行遍历
				//
				lv := tt.args.l.Get(1)
				if lv.Type() == lua.LTTable {
					lt := lv.(*lua.LTable)
					lt.ForEach(func(_ lua.LValue, value lua.LValue) {
						valT := value.(*lua.LTable)
						valT.ForEach(func(name lua.LValue, cron lua.LValue) {
							nameStr := lua.LVAsString(name)
							cronStr := lua.LVAsString(cron)
							t.Logf("show crontab:  %s : %s", nameStr, cronStr)
						})

					})
				}
			}
		})
	}
}

// 获取主机信息
func Test_provider_uname(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.uname(tt.args.l); got != tt.want {
				t.Errorf("uname() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				if lv.Type() == lua.LTTable {
					lt := lv.(*lua.LTable)
					lt.ForEach(func(key lua.LValue, value lua.LValue) {
						keyStr := lua.LVAsString(key)
						valStr := lua.LVAsString(value)
						t.Logf("key=%s val=%s", keyStr, valStr)
					})
					t.Log("uname  run ok ")
				}
			}
		})
	}
}

func Test_provider_uptime(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.uptime(tt.args.l); got != tt.want {
				t.Errorf("uptime() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				t.Log(lua.LVAsNumber(lv))
			}
		})
	}
}
