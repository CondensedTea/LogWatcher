package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	Pregame StateType = iota
	Game
)

const (
	receivedLogFile       = "received.log"
	uploaderSign          = "LogWatcher"
	logsTFURL             = "https://logs.tf/upload"
	pickupAPITemplateUrl  = "https://api.tf2pickup.%s/games"
	dryRunEnv             = "DRY_RUN"
	StartedState          = "started"
	maxSecondsAfterLaunch = 15.0
)

type StateType int

func (st *StateType) String() string {
	switch *st {
	case Pregame:
		return "Pregame"
	case Game:
		return "Game"
	default:
		return "unknown state"
	}
}

type GamesResponse struct {
	Results []struct {
		ConnectInfoVersion int    `json:"connectInfoVersion"`
		State              string `json:"state"`
		Number             int    `json:"number"`
		Map                string `json:"map"`
		Slots              []struct {
			ConnectionStatus string `json:"connectionStatus"`
			Status           string `json:"status"`
			GameClass        string `json:"gameClass"`
			Team             string `json:"team"`
			Player           string `json:"player"`
		} `json:"slots"`
		LaunchedAt       time.Time `json:"launchedAt"`
		GameServer       string    `json:"gameServer"`
		StvConnectString string    `json:"stvConnectString"`
		ID               string    `json:"id"`
		LogsUrl          string    `json:"logsUrl,omitempty"`
		Score            struct {
			Red int `json:"red"`
			Blu int `json:"blu"`
		} `json:"score,omitempty"`
		DemoUrl string `json:"demoUrl,omitempty"`
	} `json:"results"`
	ItemCount int `json:"itemCount"`
}

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	gameOver   = regexp.MustCompile(`: World triggered "Game_Over" reason "`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
	mapLoaded  = regexp.MustCompile(`: Loading map "(.+?)"`)

	dryRun = false
)

type ClientInterface interface {
	Do(r *http.Request) (*http.Response, error)
}

func init() {
	if os.Getenv(dryRunEnv) != "" {
		dryRun = true
	}
}

type LogFile struct {
	sync.Mutex
	Server   int
	Domain   string
	IP       string
	State    StateType
	channel  chan string
	buffer   bytes.Buffer
	PickupID int
	GameMap  string
	apiKey   string
}

func (lf *LogFile) Origin() string {
	return fmt.Sprintf("%s#%d", lf.Domain, lf.Server)
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
		lf.tryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			_, err := lf.buffer.WriteString(msg + "\n")
			if err != nil {
				log.WithFields(logrus.Fields{
					"server": lf.Origin(),
				}).Fatalf("Failed to write to LogFile buffer: %s", err)
			}
			if err = lf.updatePickupID(client); err != nil {
				log.WithFields(logrus.Fields{
					"server": lf.Origin(),
				}).Fatalf("Failed to get pickup id from API: %s", err)
			}
			lf.State = Game
			log.WithFields(logrus.Fields{
				"server":    lf.Origin(),
				"pickup_id": lf.PickupID,
				"map":       lf.GameMap,
			}).Infof("Succesifully parsed pickup ID")
		}
	case Game:
		_, err := lf.buffer.WriteString(msg + "\n")
		if err != nil {
			log.WithFields(logrus.Fields{
				"server": lf.Origin(),
				"state":  lf.State.String(),
			}).Fatalf("Failed to write to LogFile buffer: %s", err)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			lf.State = Pregame
			if !dryRun {
				if err = lf.uploadLogFile(client); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to upload file to logs.tf: %s", err)
				}
			} else {
				if err = saveFile(lf.buffer, receivedLogFile); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to save file to disk: %s", err)
				}
			}
			lf.flush()
		}
	}
}

func (lf *LogFile) tryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		lf.GameMap = match[1]
	}
}

func (lf *LogFile) makeMultipartMap() map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", lf.Domain, lf.PickupID))
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
		if _, err = io.Copy(fw, reader); err != nil {
			return err
		}
	}
	w.Close()

	req, err := http.NewRequest(http.MethodPost, logsTFURL, &b)
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
			return fmt.Errorf("failed to save failed log to disk, logstf response: %s err=%s", string(bodyBytes), err)
		}
		return fmt.Errorf("logs.tf returned code: %d, body: %s", res.StatusCode, string(bodyBytes))
	}
	return nil
}

func (lf *LogFile) updatePickupID(client ClientInterface) error {
	var gr GamesResponse
	url := fmt.Sprintf(pickupAPITemplateUrl, lf.Domain)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("api.tf2pickup.%s/games returned bad status: %d", lf.Domain, resp.StatusCode)
	}
	defer resp.Body.Close()

	if err = json.NewDecoder(resp.Body).Decode(&gr); err != nil {
		return err
	}

	for _, result := range gr.Results {
		if result.State == StartedState &&
			result.Map == lf.GameMap &&
			time.Since(result.LaunchedAt).Seconds() < maxSecondsAfterLaunch {
			id, err := strconv.Atoi(result.ID)
			if err != nil {
				return err
			}
			lf.PickupID = id
			break
		}
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
