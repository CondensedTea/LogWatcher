package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type State int

const (
	Pregame State = iota
	Game    State = iota
)

const (
	receivedLogFile = "received.log"
	region          = "ru"
	uploaderSign    = "LogWatcher"
	logsTFURL       = "http://logs.tf/upload"
	dryRunEnv       = "DRY_RUN"
)

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	gameOver   = regexp.MustCompile(`: World triggered "Game_Over" reason "`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
)

type ClientInterface interface {
	Do(r *http.Request) (*http.Response, error)
}

type LogFile struct {
	server  int
	region  string
	ip      string
	state   State
	channel chan string
	buffer  bytes.Buffer
	sync.Mutex
	pickupID int
	matchMap string
	apiKey   string
}

func (lf *LogFile) StartWorker() {
	client := http.Client{Timeout: 5 * time.Second}
	for msg := range lf.channel {
		lf.processLogLine(msg, &client)
	}
}

func (lf *LogFile) processLogLine(msg string, client ClientInterface) {
	lf.Lock()
	defer lf.Unlock()
	switch lf.state {
	case Pregame:
		if roundStart.MatchString(msg) {
			_, err := lf.buffer.WriteString(msg + "\n")
			if err != nil {
				log.Fatal(err)
			}
			lf.state = Game
		}
	case Game:
		_, err := lf.buffer.WriteString(msg + "\n")
		if err != nil {
			log.Fatal(err)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			lf.state = Pregame
			if os.Getenv(dryRunEnv) == "" {
				if err = lf.uploadLogFile(client); err != nil {
					log.Fatal(err)
				}
			} else {
				if err = saveFile(lf.buffer, receivedLogFile); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func (lf *LogFile) makeMultipartMap() map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", region, lf.pickupID))
	m["map"] = strings.NewReader(lf.matchMap)
	m["key"] = strings.NewReader(lf.apiKey)
	m["logfile"] = &lf.buffer
	m["uploader"] = strings.NewReader(uploaderSign)
	return m
}

func (lf *LogFile) uploadLogFile(client ClientInterface) error {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	m := lf.makeMultipartMap()
	for key, reader := range m {
		var fw io.Writer
		var err error
		if key == "logfile" {
			if fw, err = w.CreateFormFile(key, "upload.log"); err != nil {
				return err
			}
		} else {
			if fw, err = w.CreateFormField(key); err != nil {
				return err
			}
		}
		if _, err := io.Copy(fw, reader); err != nil {
			return err
		}
	}
	w.Close()

	req, err := http.NewRequest("POST", logsTFURL, &b)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(res.Body)
		if err = saveFile(lf.buffer, "logstf_failed_upload.log"); err != nil {
			return fmt.Errorf("failed to save err log after this: %s err=%s", string(bodyBytes), err)
		}
		return fmt.Errorf("logs.tf returned code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}
	return nil
}

func saveFile(buf bytes.Buffer, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(file)
	if err != nil {
		return err
	}
	err = file.Sync()
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}
