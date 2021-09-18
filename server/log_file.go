package main

import (
	"bytes"
	"os"
	"regexp"
)

const receivedLogFile = "received.log"

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
)

type LogFile struct {
	label   string
	ip      string
	state   State
	channel chan string
	buffer  bytes.Buffer
}

func (lf LogFile) StartWorker() error {
	for msg := range lf.channel {
		switch lf.state {
		case Pregame:
			if roundStart.MatchString(msg) {
				_, err := lf.buffer.WriteString(msg + "\n")
				if err != nil {
					return err
				}
				lf.state = Game
			}
		case Game:
			if !logClosed.MatchString(msg) {
				_, err := lf.buffer.WriteString(msg + "\n")
				if err != nil {
					return err
				}
			} else {
				_, err := lf.buffer.WriteString(msg + "\n")
				if err != nil {
					return err
				}
				lf.state = Pregame
				//postLogFile()
				file, err := os.Create(receivedLogFile)
				if err != nil {
					return err
				}
				_, err = lf.buffer.WriteTo(file)
				if err != nil {
					return err
				}
				file.Sync()
				file.Close()
			}
		}
	}
	return nil
}
