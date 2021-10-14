package system

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestCrontab(t *testing.T) {
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
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Crontab(tt.args.l); got != tt.want {
				t.Errorf("crontab() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestProcessOpendFiles(t *testing.T) {
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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ProcessOpendFiles(tt.args.l); got != tt.want {
				t.Errorf("processOpendFiles() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestKernelInfo(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := KernelInfo(tt.args.l); got != tt.want {
				t.Errorf("kernelInfo() = %v, want %v", got, tt.want)
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

func TestKernelModules(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// open:/proc/modules
			if got := KernelModules(tt.args.l); got != tt.want {
				t.Errorf("kernelModules() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestLast(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Last(tt.args.l); got != tt.want {
				t.Errorf("last() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestLastb(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Lastb(tt.args.l); got != tt.want {
				t.Errorf("lastb() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestLoggedInUsers(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := LoggedInUsers(tt.args.l); got != tt.want {
				t.Errorf("loggedInUsers() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

func TestShellHistory(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linuxOS", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := ShellHistory(tt.args.l); got != tt.want {
				t.Errorf("shellHistory() = %v, want %v", got, tt.want)
			} else {
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

func TestUsers(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Users(tt.args.l); got != tt.want {
				t.Errorf("users() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

// 用户名密码文件 检测
func TestShadow(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Shadow(tt.args.l); got != tt.want {
				t.Errorf("shadow() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				rangeDepth2Tabel(lv, t)
			}
		})
	}
}

// 执行sysctl 命令 返回结果
func TestSysctl(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		tt.args.l.Push(lua.LString("fs.suid_dumpable"))
		t.Run(tt.name, func(t *testing.T) {
			if got := Sysctl(tt.args.l); got != tt.want {
				t.Errorf("sysctl() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				if lv.Type() == lua.LTTable {
					lt := lv.(*lua.LTable)
					lt.ForEach(func(key lua.LValue, value lua.LValue) {
						t.Logf("key = %s val=%s", lua.LVAsString(key), lua.LVAsString(value))
					})
				}
			}
		})
	}
}

func TestProcesses(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := Processes(tt.args.l); got != tt.want {
				t.Errorf("processes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRpmList(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := RpmList(tt.args.l); got != tt.want {
				t.Errorf("rpmList() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				t.Log(lua.LVAsString(lv))
			}
		})
	}
}

func TestRpmQuery(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if got := RpmQuery(tt.args.l); got != tt.want {
				t.Errorf("rpmQuery() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(1)
				t.Log(lua.LVAsString(lv))
			}
		})
	}
}
