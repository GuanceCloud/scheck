package checker

import (
	"testing"
	"time"
)

func BenchmarkGetState(b *testing.B) {
	InitStatePool(15, 20)
	for i := 0; i < b.N; i++ {
		task := pool.getState()
		time.Sleep(time.Millisecond)
		_ = task.Ls.DoString(`print("Hello World")`)
		pool.putPool(task)
	}
}

// nolint
func TestInitStatePool(t *testing.T) {
	type args struct {
		initCap int
		totCap  int
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "eq", args: args{initCap: 10, totCap: 15}},
		{name: "more", args: args{initCap: 15, totCap: 15}},
		{name: "less", args: args{initCap: 20, totCap: 15}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InitStatePool(tt.args.initCap, tt.args.totCap)
		})
	}
}
