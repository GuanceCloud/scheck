package utils

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestHashMd5(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "winOS", args: args{l: lua.NewState()}, want: 1},
		{name: "linuxOS", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		tt.args.l.Push(lua.LString("../testdate/file.txt"))
		t.Run(tt.name, func(t *testing.T) {
			if got := HashMd5(tt.args.l); got != tt.want {
				t.Errorf("HashMd5() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(tt.args.l.GetTop())
				if lv.Type() == lua.LTString {
					t.Logf("%s", lua.LVAsString(lv))
				}
			}
		})
	}
}

func TestHashSha1(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "case2", args: args{l: lua.NewState()}, want: 1},
		{name: "case2", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		tt.args.l.Push(lua.LString("../testdate/file.txt"))
		t.Run(tt.name, func(t *testing.T) {
			if got := HashSha1(tt.args.l); got != tt.want {
				t.Errorf("HashSha1() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(tt.args.l.GetTop())
				if lv.Type() == lua.LTString {
					t.Logf("%s", lua.LVAsString(lv))
				}
			}
		})
	}
}

func TestHashSha256(t *testing.T) {
	type args struct {
		l *lua.LState
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{name: "case2", args: args{l: lua.NewState()}, want: 1},
		{name: "case2", args: args{l: lua.NewState()}, want: 1},
	}
	for _, tt := range tests {
		tt := tt
		tt.args.l.Push(lua.LString("../testdate/file.txt"))
		t.Run(tt.name, func(t *testing.T) {
			if got := HashSha256(tt.args.l); got != tt.want {
				t.Errorf("HashSha256() = %v, want %v", got, tt.want)
			} else {
				lv := tt.args.l.Get(tt.args.l.GetTop())
				if lv.Type() == lua.LTString {
					t.Logf("%s", lua.LVAsString(lv))
				}
			}
		})
	}
}
