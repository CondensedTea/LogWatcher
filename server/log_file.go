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

type StateType int

const (
	Pregame StateType = iota
	Game    StateType = iota
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
	mapLoaded  = regexp.MustCompile(`: Loading map "([a-z-_])"`)
)

type ClientInterface interface {
	Do(r *http.Request) (*http.Response, error)
}

type LogFile struct {
	Server  int
	Region  string
	IP      string
	State   StateType
	channel chan string
	buffer  bytes.Buffer
	sync.Mutex
	PickupID int
	GameMap  string
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
	switch lf.State {
	case Pregame:
		if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
			lf.GameMap = match[1]
		}
		if roundStart.MatchString(msg) {
			_, err := lf.buffer.WriteString(msg + "\n")
			if err != nil {
				log.Fatal(err)
			}
			lf.State = Game
		}
	case Game:
		_, err := lf.buffer.WriteString(msg + "\n")
		if err != nil {
			log.Fatal(err)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			lf.State = Pregame
			if os.Getenv(dryRunEnv) == "" {
				if err = lf.uploadLogFile(client); err != nil {
					log.Fatal(err)
				}
			} else {
				if err = saveFile(lf.buffer, receivedLogFile); err != nil {
					log.Fatal(err)
				}
			}
			lf.flush()
		}
	}
}

func (lf *LogFile) makeMultipartMap() map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", region, lf.PickupID))
	m["map"] = strings.NewReader(lf.GameMap)
	m["key"] = strings.NewReader(lf.apiKey)
	m["logfile"] = &lf.buffer
	m["uploader"] = strings.NewReader(uploaderSign)
	return m
}

func (lf *LogFile) flush() {
	lf.buffer = bytes.Buffer{}
	lf.PickupID = 0
	lf.GameMap = ""
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
