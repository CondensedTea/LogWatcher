package server

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"

	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
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
	uploaderSign = "LogWatcher"
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

type Server struct {
	log *logrus.Logger
	ctx context.Context
	sync.Mutex
	Server  stats.ServerInfo
	State   StateType
	Channel chan string
	buffer  bytes.Buffer
	Game    *GameInfo
	apiKey  string
	conn    *mongo.Client
}

func (s *Server) Origin() string {
	return fmt.Sprintf("%s#%d", s.Server.Domain, s.Server.ID)
}

func (s *Server) StartWorker() {
	client := http.Client{Timeout: 10 * time.Second}
	for msg := range s.Channel {
		s.processLogLine(msg, &client)
	}
}

func (s *Server) processLogLine(msg string, client requests.ClientInterface) {
	s.Lock()
	defer s.Unlock()
	switch s.State {
	case Pregame:
		s.tryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			_, err := s.buffer.WriteString(msg + "\n")
			if err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to write to Server buffer: %s", err)
			}
			if err = s.updatePickupInfo(client); err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to get pickup id from API: %s", err)
			}
			err = requests.ResolvePlayers(client, s.Server.Domain, s.Game.Players)
			if err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to resolve pickup player ids through API: %s", err)
			}
			s.State = Game
			s.log.WithFields(logrus.Fields{
				"app":       s.Origin(),
				"pickup_id": s.Game.PickupID,
				"map":       s.Game.Map,
			}).Infof("Succesifully parsed pickup ID")
		}
	case Game:
		_, err := s.buffer.WriteString(msg + "\n")
		if err != nil {
			s.log.WithFields(logrus.Fields{
				"app":   s.Origin(),
				"state": s.State.String(),
			}).Errorf("Failed to write to Server buffer: %s", err)
		}
		if err = stats.UpdatePlayerStats(msg, s.Game.Stats); err != nil {
			s.log.WithFields(logrus.Fields{
				"app":   s.Origin(),
				"state": s.State.String(),
				"msg":   msg,
			}).Errorf("Error on updating player stats: %s", err)
		}
		for k, v := range s.Game.Stats {
			s.log.Infof("STATS: %#v, %#v", k, v)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			s.State = Pregame
			payload := s.MakeMultipartMap()
			if err = requests.UploadLogFile(client, payload); err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to upload file to logs.tf: %s", err)
			}
			playersStats := stats.ExtractPlayerStats(s.Game.Stats, s.Server, s.Game.PickupID)
			if err = stats.InsertGameStats(s.ctx, s.conn, playersStats); err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to insert stats to db: %s", err)
			}
			s.flush()
		}
	}
}

func (s *Server) tryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		s.Game.Map = match[1]
	}
}

func (s *Server) flush() {
	s.buffer = bytes.Buffer{}
	s.Game.PickupID = 0
	s.Game.Map = ""
	s.Game.Stats = make(map[steamid.SID64]*stats.PlayerStats)
}

func (s *Server) MakeMultipartMap() map[string]io.Reader {
	m := make(map[string]io.Reader)
	m["title"] = strings.NewReader(fmt.Sprintf("tf2pickup.%s #%d", s.Server.Domain, s.Game.PickupID))
	m["map"] = strings.NewReader(s.Game.Map)
	m["key"] = strings.NewReader(s.apiKey)
	m["logfile"] = &s.buffer
	m["uploader"] = strings.NewReader(uploaderSign)
	return m
}

func MakeRouterMap(hosts []config.Client, apiKey string, conn *mongo.Client, log *logrus.Logger) map[string]*Server {
	serverMap := make(map[string]*Server)
	for _, h := range hosts {
		lf := &Server{
			log: log,
			ctx: context.Background(),
			Server: stats.ServerInfo{
				ID:     h.Server,
				Domain: h.Domain,
				IP:     h.Address,
			},
			State:   Pregame,
			Channel: make(chan string),
			Game: &GameInfo{
				Players: make([]*requests.PickupPlayer, 0),
				Stats:   make(map[steamid.SID64]*stats.PlayerStats),
			},
			apiKey: apiKey,
			conn:   conn,
		}
		go lf.StartWorker()
		serverMap[h.Address] = lf
		lf.log.Infof("Started worker for %s#%d with host %s", lf.Server.Domain, lf.Server.ID, lf.Server.IP)
	}
	return serverMap
}

func (s *Server) updatePickupInfo(client requests.ClientInterface) error {
	gr, err := requests.GetPickupGames(client, s.Server.Domain)
	if err != nil {
		return err
	}
	for _, result := range gr.Results {
		if result.State == StartedState && result.Map == s.Game.Map {
			players := make([]*requests.PickupPlayer, 0)
			for _, player := range result.Slots {
				p := &requests.PickupPlayer{PlayerID: player.Player, Class: player.GameClass}
				players = append(players, p)
			}
			s.Game.Players = players
			s.Game.PickupID = result.Number
			break
		}
	}
	return nil
}

type GameInfo struct {
	PickupID int
	Map      string
	Players  []*requests.PickupPlayer
	Stats    map[steamid.SID64]*stats.PlayerStats
}
