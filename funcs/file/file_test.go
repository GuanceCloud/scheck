package file

import (
	"runtime"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

const (
	winOS   = "windows"
	linuxOS = "linux"
)

func Test_FileExist(t *testing.T) {
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
		tt := tt
		if tt.name == winOS {
			tt.args.l.Push(lua.LString("C:\\users"))
		}
		if tt.name == linuxOS {
			tt.args.l.Push(lua.LString("/usr/local/bin"))
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := Exist(tt.args.l); got != tt.want {
				t.Errorf("fileExist() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(tt.args.l.GetTop())
				if lv.Type() == lua.LTBool {
					t.Log(lua.LVAsBool(lv))
				}
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
	for _, test := range tests {
		test := test
		if runtime.GOOS == winOS && test.name == winOS {
			test.args.l.Push(lua.LString("C:\\Windows\\System32\\drivers\\etc\\hosts"))
			t.Run(test.name, func(t *testing.T) {
				if got := Info(test.args.l); got != test.want {
					t.Errorf("fileInfo() = %v, want %v", got, test.want)
				} else {
					lv := test.args.l.Get(test.args.l.GetTop())
					if lv.Type() == lua.LTTable {
						lt := lv.(*lua.LTable)
						lt.ForEach(func(key lua.LValue, value lua.LValue) {
							keyStr := lua.LVAsString(key)
							valStr := lua.LVAsString(value)
							t.Logf("key=%s val=%s", keyStr, valStr)
						})
					} else {
						t.Log(lua.LVAsString(lv))
					}
				}
			})
		}
		if runtime.GOOS == linuxOS && test.name == linuxOS {
			test.args.l.Push(lua.LString("/etc/hosts"))
			t.Run(test.name, func(t *testing.T) {
				if got := Info(test.args.l); got != test.want {
					t.Errorf("fileInfo() = %v, want %v", got, test.want)
				} else {
					lv := test.args.l.Get(test.args.l.GetTop())
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

func TestGrep(t *testing.T) {
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
		tt := tt
		tt.args.l.Push(lua.LString("/etc/sudoers"))
		tt.args.l.Push(lua.LString("^\\s*Defaults\\s+([^#]\\S+,\\s*)?use_pty\\b"))
		tt.args.l.Push(lua.LString("-Ei"))
		t.Run(tt.name, func(t *testing.T) {
			if got := Grep(tt.args.l); got != tt.want {
				t.Errorf("grep() = %v, want %v", got, tt.want)
			} else {
				// 返回两个返回值 string，error
				lv := tt.args.l.Get(1)
				err := lua.LVAsString(lv)
				if err != "" {
					t.Logf("err =%s", err)
					return
				}
				lv2 := tt.args.l.Get(2)
				msg := lua.LVAsString(lv2)
				t.Log(msg)
			}
		})
	}
}
