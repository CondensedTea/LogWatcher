package router

import (
	"LogWatcher/pkg/config"

	"LogWatcher/pkg/server"
	"net"
	"regexp"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

type Router struct {
	address   *net.UDPAddr
	routerMap map[string]*server.Server
	log       *logrus.Logger
}

func NewRouter(cfg *config.Config, conn *mongo.Client, log *logrus.Logger) (*Router, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		return nil, err
	}
	m := server.MakeRouterMap(cfg.Clients, cfg.Server.APIKey, conn, log)

	return &Router{
		address:   udpAddr,
		routerMap: m,
		log:       log,
	}, nil
}

func (s *Router) Listen() {
	conn, err := net.ListenUDP("udp", s.address)
	if err != nil {
		s.log.Fatalf("failed to listen UDP port: %s", err)
	}
	s.log.Infof("LogWatcher is listening on %s", s.address.String())
	for {
		message := make([]byte, 1024)
		msgLen, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			s.log.Errorf("Failed to read from UDP socket: %s", err)
			return
		}
		cleanMsg := logLineRegexp.FindString(string(message[:msgLen]))

		lf, ok := s.routerMap[clientAddr.String()]
		if !ok {
			s.log.WithFields(logrus.Fields{
				"address": clientAddr.String(),
				"app":     "unknown",
			}).Debugf(cleanMsg)
			continue
		}
		s.log.WithFields(logrus.Fields{
			"app":       lf.Origin(),
			"state":     lf.State.String(),
			"pickup_id": lf.Game.PickupID,
		}).Infof(cleanMsg)
		lf.Channel <- cleanMsg
	}
}
