package main

import (
	"reflect"
	"testing"
)

func Test_loadConfig(t *testing.T) {
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
					{Server: 1, Region: "<region>", Address: "<ip>"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loadConfig(tt.args.path)
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
