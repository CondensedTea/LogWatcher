package main

import (
	"bytes"
	"log"
	"os"
	"regexp"
)

type State int

const (
	Pregame State = iota
	Game    State = iota
)

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
)

type Upload struct {
	Title string
	Map   string
	Key   string
}

type Client struct {
	name    string
	address string
}

type LogFile struct {
	state   State
	channel chan string
	buffer  bytes.Buffer
}

//func postLogFile(fields map[string]string, buffer bytes.Buffer) {
//	writer := multipart.NewWriter(&buffer)
//	defer writer.Close()
//	writer.CreateFormFile()
//}

func (lf LogFile) StartWorker() {
	go func() {
		for msg := range lf.channel {
			switch lf.state {
			case Pregame:
				if roundStart.MatchString(msg) {
					lf.state = Game
				}
			case Game:
				if !logClosed.MatchString(msg) {
					lf.buffer.WriteString(msg)
				}
				lf.state = Pregame
				//postLogFile()
				file, err := os.Open("test.log")
				if err != nil {
					log.Fatalf("Failed to open logfile: %s", err)
				}
				_, err = lf.buffer.WriteTo(file)
				if err != nil {
					log.Fatalf("Failed to write to file: %s", err)
				}
				file.Close()
			}
		}
	}()
}

type Config struct {
	Server struct {
		Host string
		Port string
	}
	Clients []Client
}
