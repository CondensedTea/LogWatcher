package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/leighmacdonald/steamid/steamid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	Pregame StateType = iota
	Game
)

const (
	StartedState = "started"
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

type PickupPlayer struct {
	PlayerID string `bson:"player_id"`
	Class    string
	SteamID  string `bson:"steam_id"`
}

type LogFile struct {
	ctx context.Context
	sync.Mutex
	Server  ServerInfo
	State   StateType
	channel chan string
	buffer  bytes.Buffer
	Game    *GameInfo
	apiKey  string
	conn    *mongo.Client
}

func (lf *LogFile) Origin() string {
	return fmt.Sprintf("%s#%d", lf.Server.Domain, lf.Server.ID)
}

func (lf *LogFile) StartWorker() {
	client := http.Client{Timeout: 10 * time.Second}
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
				log.WithFields(logrus.Fields{"server": lf.Origin()}).
					Errorf("Failed to write to LogFile buffer: %s", err)
			}
			if err = lf.updatePickupInfo(client); err != nil {
				log.WithFields(logrus.Fields{"server": lf.Origin()}).
					Errorf("Failed to get pickup id from API: %s", err)
			}
			if err = lf.resolvePlayers(client); err != nil {
				log.WithFields(logrus.Fields{"server": lf.Origin()}).
					Errorf("Failed to resolve pickup player ids through API: %s", err)
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
			}).Errorf("Failed to write to LogFile buffer: %s", err)
		}
		if err = lf.Game.updatePlayerStats(msg); err != nil {
			log.WithFields(logrus.Fields{
				"server": lf.Origin(),
				"state":  lf.State.String(),
				"msg":    msg,
			}).Errorf("Error on updating player stats: %s", err)
		}
		for k, v := range lf.Game.Stats {
			log.Infof("STATS: %#v, %#v", k, v)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			lf.State = Pregame
			if err = lf.uploadLogFile(client); err != nil {
				log.WithFields(logrus.Fields{"server": lf.Origin()}).
					Errorf("Failed to upload file to logs.tf: %s", err)
			}
			stats := lf.ExtractPlayerStats()
			if err = lf.insertGameStats(stats); err != nil {
				log.WithFields(logrus.Fields{"server": lf.Origin()}).
					Errorf("Failed to insert stats to db: %s", err)
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
