package server

import (
	"LogWatcher/pkg/config"
	"bytes"
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
	buffer bytes.Buffer
}

func NewLogFile(client config.Client) *LogFile {
	return &LogFile{
		name:   fmt.Sprintf("%s#%d", client.Domain, client.Server),
		buffer: bytes.Buffer{},
	}
}

func (s *LogFile) Name() string {
	return s.name
}

func (s *LogFile) WriteLine(msg string) {
	s.buffer.WriteString(msg + "\n")
}

func (s *LogFile) Buffer() bytes.Buffer {
	return s.buffer
}

func (s *LogFile) FlushBuffer() {
	s.buffer.Reset()
}
