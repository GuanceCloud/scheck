// nolint
package system

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func Test_provider_grep(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "linux1", args: args{l: lua.NewState()}, want: 2},
		{name: "linux2", args: args{l: lua.NewState()}, want: 2},
	}
	for _, tt := range tests {
		tt.args.l.Push(lua.LString("/etc/sudoers"))
		tt.args.l.Push(lua.LString("^\\s*Defaults\\s+([^#]\\S+,\\s*)?use_pty\\b"))
		tt.args.l.Push(lua.LString("-Ei"))

		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.grep(tt.args.l); got != tt.want {
				t.Errorf("grep() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_crontab(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.crontab(tt.args.l); got != tt.want {
				t.Errorf("crontab() = %v, want %v", got, tt.want)
			} else {
				// 返回的是二维table 把value 转成table再进行遍历
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

func Test_provider_processOpendFiles(t *testing.T) {
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
			if got := p.processOpendFiles(tt.args.l); got != tt.want {
				t.Errorf("processOpendFiles() = %v, want %v", got, tt.want)
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

func Test_provider_kernelInfo(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.kernelInfo(tt.args.l); got != tt.want {
				t.Errorf("kernelInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_kernelModules(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.kernelModules(tt.args.l); got != tt.want {
				t.Errorf("kernelModules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_last(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.last(tt.args.l); got != tt.want {
				t.Errorf("last() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_lastb(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.lastb(tt.args.l); got != tt.want {
				t.Errorf("lastb() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_loggedInUsers(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.loggedInUsers(tt.args.l); got != tt.want {
				t.Errorf("loggedInUsers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_shellHistory(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: linuxOS, args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.shellHistory(tt.args.l); got != tt.want {
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

func Test_provider_users(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.users(tt.args.l); got != tt.want {
				t.Errorf("users() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 用户名密码文件 检测
func Test_provider_shadow(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.shadow(tt.args.l); got != tt.want {
				t.Errorf("shadow() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 执行sysctl 命令 返回结果
func Test_provider_sysctl(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.sysctl(tt.args.l); got != tt.want {
				t.Errorf("sysctl() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_processes(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.processes(tt.args.l); got != tt.want {
				t.Errorf("processes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_rpmList(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.rpmList(tt.args.l); got != tt.want {
				t.Errorf("rpmList() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_provider_rpmQuery(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.rpmQuery(tt.args.l); got != tt.want {
				t.Errorf("rpmQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 获取ulimit信息
func Test_provider_ulimitInfo(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &provider{}
			if got := p.ulimitInfo(tt.args.l); got != tt.want {
				t.Errorf("ulimitInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
