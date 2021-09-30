package main

import (
	"context"
	"net"
	"regexp"

	"github.com/leighmacdonald/steamid/steamid"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

type Server struct {
	address    *net.UDPAddr
	addressMap map[string]*LogFile
}

func makeAddressMap(hosts []Client, dryRun bool, apiKey string, conn *mongo.Client) map[string]*LogFile {
	logsDict := make(map[string]*LogFile)
	for _, h := range hosts {
		s := ServerInfo{
			ID:     h.Server,
			Domain: h.Domain,
			IP:     h.Address,
		}
		gi := &GameInfo{
			Players: make([]PickupPlayer, 0),
			Stats:   make(map[steamid.SID64]*PlayerStats),
		}
		lf := &LogFile{
			ctx:     context.Background(),
			Server:  s,
			State:   Pregame,
			channel: make(chan string),
			Game:    gi,
			apiKey:  apiKey,
			dryRun:  dryRun,
			conn:    conn,
		}
		go lf.StartWorker()
		logsDict[h.Address] = lf
		log.Infof("Started worker for %s#%d with host %s", lf.Server.Domain, lf.Server.ID, lf.Server.IP)
	}
	return logsDict
}

func NewServer(cfg *Config, conn *mongo.Client) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		return nil, err
	}
	m := makeAddressMap(cfg.Clients, cfg.Server.DryRun, cfg.Server.APIKey, conn)

	return &Server{
		address:    udpAddr,
		addressMap: m,
	}, nil
}

func (s *Server) Listen() {
	conn, err := net.ListenUDP("udp", s.address)
	if err != nil {
		log.Fatalf("failed to listen UDP port: %s", err)
	}
	log.Infof("LogWatcher is listening on %s", s.address.String())
	for {
		message := make([]byte, 1024)
		msgLen, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Fatalf("Failed to read from UDP socket: %s", err)
		}
		cleanMsg := logLineRegexp.FindString(string(message[:msgLen]))

		lf, ok := s.addressMap[clientAddr.String()]
		if !ok {
			log.WithFields(logrus.Fields{
				"address": clientAddr.String(),
				"server":  "unknown",
			}).Debugf(cleanMsg)
			continue
		}
		log.WithFields(logrus.Fields{
			"server":    lf.Origin(),
			"state":     lf.State.String(),
			"pickup_id": lf.Game.PickupID,
		}).Infof(cleanMsg)
		lf.channel <- cleanMsg
	}
}
