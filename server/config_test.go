package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_LoadConfig(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    *Config
		wantErr bool
	}{
		{
			name: "default",
			args: args{"../e2e/e2e_config.yaml"},
			want: &Config{
				Server: struct {
					Host   string `yaml:"Host"`
					APIKey string `yaml:"APIKey"`
					DSN    string `yaml:"DSN"`
				}{Host: "localhost:27100", APIKey: "fake", DSN: "fake"},
				Clients: []Client{
					{Server: 1, Domain: "test", Address: "127.0.0.1:27150"},
				},
			},
			wantErr: false,
		},
		{
			name:    "no file",
			args:    args{"../fake-e2e_config.yaml"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "bad yaml",
			args:    args{"../Dockerfile"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !cmp.Equal(got, tt.want) {
				t.Errorf("loadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
