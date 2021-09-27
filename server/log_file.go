package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/jackc/pgx"
	"github.com/leighmacdonald/steamid/steamid"
	"github.com/sirupsen/logrus"
)

const (
	Pregame StateType = iota
	Game
)

const (
	receivedLogFile = "received.log"
	StartedState    = "started"
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

var (
	roundStart = regexp.MustCompile(`: World triggered "Round_Start"`)
	gameOver   = regexp.MustCompile(`: World triggered "Game_Over" reason "`)
	logClosed  = regexp.MustCompile(`: Log file closed.`)
	mapLoaded  = regexp.MustCompile(`: Loading map "(.+?)"`)
)

type ServerInfo struct {
	ID     int
	Domain string
	IP     string
}

type GameInfo struct {
	PickupID int
	Map      string
	Players  []PickupPlayer
	Stats    map[steamid.SID64]*PlayerStats
}

type LogFile struct {
	sync.Mutex
	Server  ServerInfo
	State   StateType
	channel chan string
	buffer  bytes.Buffer
	Game    *GameInfo
	apiKey  string
	dryRun  bool
	conn    *pgx.Conn
}

func (lf *LogFile) Origin() string {
	return fmt.Sprintf("%s#%d", lf.Server.Domain, lf.Server.ID)
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
			if !lf.dryRun {
				if err = lf.updatePickupInfo(client); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to get pickup id from API: %s", err)
				}
				if err = lf.resolvePlayers(client); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to resolve pickup player ids through API: %s", err)
				}
			}
			lf.State = Game
			log.WithFields(logrus.Fields{
				"server":    lf.Origin(),
				"pickup_id": lf.Game.PickupID,
				"map":       lf.Game.Map,
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
		if err = lf.Game.updatePlayerStats(msg); err != nil {
			log.WithFields(logrus.Fields{
				"server": lf.Origin(),
				"state":  lf.State.String(),
				"msg":    msg,
			}).Errorf("Error on updating player stats: %s", err)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			lf.State = Pregame
			if !lf.dryRun {
				if err = lf.uploadLogFile(client); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to upload file to logs.tf: %s", err)
				}
				if err = lf.insertPlayerStats(); err != nil {
					log.WithFields(logrus.Fields{
						"server": lf.Origin(),
					}).Fatalf("Failed to insert stats to db: %s", err)
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
		lf.Game.Map = match[1]
	}
}

func (lf *LogFile) flush() {
	lf.buffer = bytes.Buffer{}
	lf.Game.PickupID = 0
	lf.Game.Map = ""
	lf.Game.Stats = make(map[steamid.SID64]*PlayerStats)
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
