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
			args: args{"/Users/atyshkevich/LogWatcher/config.yaml"},
			want: &Config{
				Server: struct {
					Host string `yaml:"Host"`
				}{Host: "0.0.0.0:27100"},
				Clients: []Client{
					{Name: "ru.1", Address: "46.174.50.10"},
					{Name: "test", Address: "::1"},
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
