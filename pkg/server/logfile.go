package server

import (
	"LogWatcher/pkg/config"
	"bytes"
	"context"
	"fmt"
)

type LogFiler interface {
	Name() string
	WriteLine(msg string)
	Buffer() bytes.Buffer
	FlushBuffer()
}

type LogFile struct {
	name   string
	ctx    context.Context
	buffer bytes.Buffer
}

func NewLogFile(client config.Client) *LogFile {
	return &LogFile{
		name:   fmt.Sprintf("%s#%d", client.Domain, client.Server),
		ctx:    context.Background(),
		buffer: bytes.Buffer{},
	}
}

func (s *LogFile) Name() string {
	return s.name
}

//func (s *LogFile) Channel() chan string {
//	return s.channel
//}

func (s *LogFile) WriteLine(msg string) {
	s.buffer.WriteString(msg + "\n")
}

func (s *LogFile) Buffer() bytes.Buffer {
	return s.buffer
}

func (s *LogFile) FlushBuffer() {
	s.buffer = bytes.Buffer{}
}
