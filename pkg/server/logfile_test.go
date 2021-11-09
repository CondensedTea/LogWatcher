package server

import (
	"LogWatcher/pkg/config"
	"bytes"
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
				buffer: &tt.fields.buffer,
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

	file := &LogFile{
		buffer: &buf,
	}
	file.FlushBuffer()
	if length := file.buffer.Len(); length > 0 {
		t.Errorf("buffer length after FlushBuffer() is %v, wanted 0", length)
	}
}

func TestLogFile_WriteLine(t *testing.T) {
	msg := "test"
	buf := &bytes.Buffer{}

	file := &LogFile{
		buffer: buf,
	}
	file.WriteLine(msg)
	text := file.buffer.String()
	if text != msg+"\n" {
		t.Errorf("WriteLine() got = %v, want = %v", text, msg)
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
	host := config.Client{
		Server: 1,
		Domain: "test",
	}
	want := &LogFile{
		name: "test#1",
	}
	result := NewLogFile(host)
	if !cmp.Equal(result.name, want.name) {
		t.Errorf("NewLogFile() = %v, want %v", result, want)
	}
}
