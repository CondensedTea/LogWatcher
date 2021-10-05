package server

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/stats"
	"encoding/json"

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
	StartedState         = "started"
	uploaderSignTemplate = "LogWatcher (%s)"
)

var Version = "testing"

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
	timeStamp  = regexp.MustCompile(`\d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}`)
)

type GameInfo struct {
	PickupID    int
	Map         string
	Players     []*stats.PickupPlayer
	Stats       map[steamid.SID64]*stats.PlayerStats
	LaunchedAt  time.Time
	MatchLength time.Duration
}

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

func (s *Server) processLogLine(msg string, client requests.HTTPDoer) {
	s.Lock()
	defer s.Unlock()
	switch s.State {
	case Pregame:
		s.tryParseGameMap(msg)
		if roundStart.MatchString(msg) {
			s.Game.LaunchedAt = parseTimeStamp(msg)
			s.buffer.WriteString(msg + "\n")
			if err := s.updatePickupInfo(client); err != nil {
				s.log.WithFields(logrus.Fields{"server": s.Origin()}).
					Errorf("Failed to get pickup id from API: %s", err)
			}
			err := requests.ResolvePlayers(client, s.Server.Domain, s.Game.Players)
			if err != nil {
				s.log.WithFields(logrus.Fields{"server": s.Origin()}).
					Errorf("Failed to resolve pickup player ids through API: %s", err)
			}
			s.State = Game
			s.log.WithFields(logrus.Fields{
				"server":    s.Origin(),
				"pickup_id": s.Game.PickupID,
				"map":       s.Game.Map,
			}).Infof("Pickup has started")
		}
	case Game:
		s.buffer.WriteString(msg + "\n")
		if err := stats.UpdateStatsMap(msg, s.Game.Stats); err != nil {
			s.log.WithFields(logrus.Fields{
				"server": s.Origin(),
				"state":  s.State.String(),
				"msg":    msg,
			}).Errorf("Error on updating player stats: %s", err)
		}
		if logClosed.MatchString(msg) || gameOver.MatchString(msg) {
			s.State = Pregame
			ts := parseTimeStamp(msg)
			s.Game.MatchLength = ts.Sub(s.Game.LaunchedAt)
			payload := s.MakeMultipartMap()
			if err := requests.UploadLogFile(client, payload); err != nil {
				s.log.WithFields(logrus.Fields{"server": s.Origin()}).
					Errorf("Failed to upload file to logs.tf: %s", err)
			}
			playersJSON, _ := json.Marshal(s.Game.Players)
			gameStatsJSON, _ := json.Marshal(s.Game.Stats)
			s.log.WithFields(logrus.Fields{
				"players":   string(playersJSON),
				"gameStats": string(gameStatsJSON),
			}).Info("Preparing to extract player stats")
			playersStats := stats.ExtractPlayerStats(s.Game.Players, s.Game.Stats, s.Server, s.Game.PickupID, s.Game.MatchLength)
			if err := stats.InsertGameStats(s.ctx, s.conn, playersStats); err != nil {
				s.log.WithFields(logrus.Fields{"app": s.Origin()}).
					Errorf("Failed to insert stats to db: %s", err)
			}
			s.log.WithFields(logrus.Fields{
				"server":    s.Origin(),
				"pickup_id": s.Game.PickupID,
				"map":       s.Game.Map,
			}).Info("Pickup has ended")
			s.flush()
		}
	}
}

// tryParseGameMap tries to find "Loading map" with regexp in message
// and sets it to Server.Game.Map if succeed
func (s *Server) tryParseGameMap(msg string) {
	if match := mapLoaded.FindStringSubmatch(msg); len(match) > 0 {
		s.Game.Map = match[1]
	}
}

// flush is used to empty all info for the game
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
	m["uploader"] = strings.NewReader(fmt.Sprintf(uploaderSignTemplate, Version))
	return m
}

func MakeRouterMap(hosts []config.Client, apiKey string, conn *mongo.Client, log *logrus.Logger) map[string]*Server {
	serverMap := make(map[string]*Server)
	for _, h := range hosts {
		s := &Server{
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
				Players: make([]*stats.PickupPlayer, 0),
				Stats:   make(map[steamid.SID64]*stats.PlayerStats),
			},
			apiKey: apiKey,
			conn:   conn,
		}
		go s.StartWorker()
		serverMap[h.Address] = s
		s.log.Infof("Started worker for %s#%d with host %s", s.Server.Domain, s.Server.ID, s.Server.IP)
	}
	return serverMap
}

// updatePickupInfo is used for finding current game on tf2pickup API
// and loading to Server list of its players and pickup ID
func (s *Server) updatePickupInfo(client requests.HTTPDoer) error {
	gr, err := requests.GetPickupGames(client, s.Server.Domain)
	if err != nil {
		return err
	}
	s.log.Debugf("Got list of games from pickup API with length of %d", gr.ItemCount)
	for _, game := range gr.Results {
		if game.State == StartedState && game.Map == s.Game.Map {
			players := make([]*stats.PickupPlayer, 0)
			s.log.WithFields(logrus.Fields{
				"len of slots": len(game.Slots),
			}).Debugf("Processing matched result")
			for _, player := range game.Slots {
				s.log.WithFields(logrus.Fields{
					"id":    player.Player,
					"class": player.GameClass,
					"team":  player.Team,
				}).Debugf("Processing result's player")
				p := &stats.PickupPlayer{PlayerID: player.Player, Class: player.GameClass, Team: player.Team}
				players = append(players, p)
			}
			s.Game.Players = players
			s.Game.PickupID = game.Number
			break
		}
	}
	return nil
}

func parseTimeStamp(msg string) time.Time {
	match := timeStamp.FindString(msg)
	t, _ := time.Parse(`01/2/2006 - 15:04:05`, match) // err is always nil
	return t
}
