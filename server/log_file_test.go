package main

import (
	"bytes"
	"os"
	"testing"
)

func Test_saveFile(t *testing.T) {
	dir := t.TempDir() + "test.file"
	type args struct {
		buf  bytes.Buffer
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				buf:  bytes.Buffer{},
				path: dir,
			},
			want:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt.args.buf.WriteString(tt.want)
		t.Run(tt.name, func(t *testing.T) {
			if err := saveFile(tt.args.buf, tt.args.path); (err != nil) != tt.wantErr {
				t.Errorf("saveFile() error = %v, wantErr %v", err, tt.wantErr)
			}
			lines, err := os.ReadFile(dir)
			if err != nil {
				t.Errorf("saveFile() read file error = %v", err)
			}
			if string(lines) != tt.want {
				t.Errorf("saveFile() file contents error got = %v, want = %v", string(lines), tt.want)
			}
		})
	}
}
