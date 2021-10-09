package server

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	gameOver   = regexp.MustCompile(`: World triggered "Game_Over" reason "`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
)

type LogFiler interface {
	Lock()
	Unlock()
	Name() string
	State() StateType
	SetState(state StateType)
	Channel() chan string
	GetConn() *mongo.Client
	WriteLine(msg string)
	Buffer() bytes.Buffer
	FlushBuffer()
}

type LogFile struct {
	name string
	log  *logrus.Logger
	ctx  context.Context
	sync.Mutex
	state   StateType
	channel chan string
	buffer  bytes.Buffer
	conn    *mongo.Client
}

func NewLogFile(log *logrus.Logger, conn *mongo.Client, domain string, id int) *LogFile {
	return &LogFile{
		name:    fmt.Sprintf("%s#%d", domain, id),
		log:     log,
		ctx:     context.Background(),
		state:   Pregame,
		buffer:  bytes.Buffer{},
		channel: make(chan string),
		conn:    conn,
	}
}

func (s *LogFile) Name() string {
	return s.name
}

func (s *LogFile) State() StateType {
	return s.state
}

func (s *LogFile) SetState(state StateType) {
	s.state = state
}

func (s *LogFile) Channel() chan string {
	return s.channel
}

func (s *LogFile) GetConn() *mongo.Client {
	return s.conn
}

func (s *LogFile) WriteLine(msg string) {
	s.buffer.WriteString(msg + "\n")
}

func (s *LogFile) Buffer() bytes.Buffer {
	return s.buffer
}

func (s *LogFile) FlushBuffer() {
	s.buffer = bytes.Buffer{}
}