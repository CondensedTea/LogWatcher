package server

import (
	"bytes"
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
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

func TestLogFile_Channel(t *testing.T) {
	ch := make(chan string)
	type fields struct {
		channel chan string
	}
	tests := []struct {
		name   string
		fields fields
		want   chan string
	}{
		{
			name: "default",
			fields: fields{
				channel: ch,
			},
			want: ch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				channel: tt.fields.channel,
			}
			if got := s.Channel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channel() = %v, want %v", got, tt.want)
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

func TestLogFile_SetState(t *testing.T) {
	type fields struct {
		state StateType
	}
	type args struct {
		state StateType
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   StateType
	}{
		{
			name:   "default",
			fields: fields{state: Pregame},
			args:   args{state: Game},
			want:   Game,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				state: tt.fields.state,
			}
			s.SetState(tt.args.state)
			if !reflect.DeepEqual(s.state, tt.want) {
				t.Errorf("SetState(), got = %v, want %v", tt.fields.state, tt.want)
			}
		})
	}
}

func TestLogFile_State(t *testing.T) {
	type fields struct {
		state StateType
	}
	tests := []struct {
		name   string
		fields fields
		want   StateType
	}{
		{
			name:   "default",
			fields: fields{state: Pregame},
			want:   Pregame,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &LogFile{
				state: tt.fields.state,
			}
			if got := s.State(); got != tt.want {
				t.Errorf("State() = %v, want %v", got, tt.want)
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
		log    *logrus.Logger
		domain string
		id     int
	}
	tests := []struct {
		name string
		args args
		want *LogFile
	}{
		{
			name: "default",
			args: args{
				domain: "test",
				id:     1,
			},
			want: &LogFile{
				name: "test#1",
				ctx:  ctx,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewLogFile(tt.args.log, tt.args.domain, tt.args.id)
			if !cmp.Equal(tt.want.name, got.name) {
				t.Errorf("NewLogFile() = %v, want %v", got, tt.want)
			}
		})
	}
}
