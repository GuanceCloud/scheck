package output

import (
	"testing"
)

func Test_newDatakitWriter(t *testing.T) {
	type args struct {
		httpURL    string
		maxPending int
	}
	tests := []struct {
		name string
		args args
		want *DatakitWriter
	}{
		{name: "case1", args: args{httpURL: "http://120.0.0.1:9529/v1/write/security", maxPending: 5}, want: &DatakitWriter{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDatakitWriter(tt.args.httpURL, tt.args.maxPending); got == nil {
				t.Errorf("newDatakitWriter() = %v, want %v", got, tt.want)
			} else {
				t.Logf("%s :%s", tt.name, got.httpURL)
			}
		})
	}
}
