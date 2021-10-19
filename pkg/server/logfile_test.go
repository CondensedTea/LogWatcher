package server

import (
	"LogWatcher/pkg/config"
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLogFile_Buffer(t *testing.T) {
	buf := bytes.Buffer{}
	type fields struct {
		buffer bytes.Buffer
	}
	tests := []struct {
		name   string
		fields fields
		want   bytes.Buffer
	}{
		{
			name:   "default",
			fields: fields{buffer: buf},
			want:   buf,
		},
		{
			name:   "sigsegv",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				buffer: tt.fields.buffer,
			}
			if got := s.Buffer(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Buffer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogFile_FlushBuffer(t *testing.T) {
	buf := bytes.Buffer{}
	buf.WriteString("hello")
	type fields struct {
		buffer bytes.Buffer
	}
	tests := []struct {
		name   string
		fields fields
		want   bytes.Buffer
	}{
		{
			name:   "default",
			fields: fields{buffer: buf},
			want:   bytes.Buffer{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				buffer: tt.fields.buffer,
			}
			s.FlushBuffer()
			if !reflect.DeepEqual(s.buffer, tt.want) {
				t.Errorf("FlushBuffer(), got = %v, want %v", tt.fields.buffer, tt.want)
			}
		})
	}
}

func TestLogFile_WriteLine(t *testing.T) {
	originalBuf := bytes.Buffer{}
	updatedBuf := bytes.Buffer{}
	updatedBuf.WriteString("test" + "\n")
	type fields struct {
		buffer bytes.Buffer
	}
	type args struct {
		msg string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bytes.Buffer
	}{
		{
			name:   "default",
			fields: fields{buffer: originalBuf},
			args:   args{msg: "test"},
			want:   updatedBuf,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				buffer: tt.fields.buffer,
			}
			s.WriteLine(tt.args.msg)
			if !reflect.DeepEqual(s.buffer, tt.want) {
				t.Errorf("WriteLine(), got = %v, want %v", s.buffer, tt.want)
			}
		})
	}
}

func TestLogFile_Name(t *testing.T) {
	type fields struct {
		name string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "default",
			fields: fields{name: "test"},
			want:   "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				name: tt.fields.name,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLogFile(t *testing.T) {
	ctx := context.Background()
	type args struct {
		host config.Client
	}
	tests := []struct {
		name string
		args args
		want *LogFile
	}{
		{
			name: "default",
			args: args{
				host: config.Client{
					Server: 1,
					Domain: "test",
				},
			},
			want: &LogFile{
				name: "test#1",
				ctx:  ctx,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLogFile(tt.args.host)
			if !cmp.Equal(tt.want.name, got.name) {
				t.Errorf("NewLogFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
