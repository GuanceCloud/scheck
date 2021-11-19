package service

import (
	"testing"

	"github.com/kardianos/service"
)

// nolint
func TestNewService(t *testing.T) {
	tests := []struct {
		name    string
		want    service.Service
		wantErr bool
	}{
		{name: "case", want: nil, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewService()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewService() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewService() got = %v, want %v", got, tt.want)
			}
		})
	}
}
