package config

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

const (
	testCfgPath = `test_config.yaml`
	fakeCfgPath = `fake_config.yaml`
	badCfgPath  = `bad_config.yaml`
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
			args: args{testCfgPath},
			want: &Config{
				Server: Server{
					Host:            "localhost:27100",
					APIKey:          "apiKey",
					DSN:             "dsn",
					MongoDatabase:   "db",
					MongoCollection: "collection",
				},
				Clients: []Client{
					{Server: 1, Domain: "test", Address: "127.0.0.1:27150"},
				},
			},
			wantErr: false,
		},
		{
			name:    "no file",
			args:    args{fakeCfgPath},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "bad yaml",
			args:    args{badCfgPath},
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
