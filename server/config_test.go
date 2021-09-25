package main

import (
	"reflect"
	"testing"
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
			args: args{"../config.template.yaml"},
			want: &Config{
				Server: struct {
					Host   string `yaml:"Host"`
					APIKey string `yaml:"APIKey"`
				}{Host: "<host>:<port>", APIKey: "<logstf-api-key>"},
				Clients: []Client{
					{Server: 1, Domain: "<your-domain>", Address: "<ip>:<port>"},
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("loadConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
