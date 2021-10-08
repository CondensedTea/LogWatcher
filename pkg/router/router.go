package router

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stats"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

type Router struct {
	address   *net.UDPAddr
	routerMap map[string]*server.LogFile
	log       *logrus.Logger
}

func NewRouter(cfg *config.Config, conn *mongo.Client, log *logrus.Logger) (*Router, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		return nil, err
	}
	m := MakeRouterMap(cfg.Clients, cfg.Server.APIKey, conn, log)
	return &Router{
		address:   udpAddr,
		routerMap: m,
		log:       log,
	}, nil
}

func (r *Router) Listen() {
	conn, err := net.ListenUDP("udp", r.address)
	if err != nil {
		r.log.Fatalf("failed to listen UDP port: %s", err)
	}
	r.log.Infof("LogWatcher is listening on %s", r.address.String())
	for {
		message := make([]byte, 1024)
		msgLen, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			r.log.Errorf("Failed to read from UDP socket: %s", err)
			return
		}
		cleanMsg := logLineRegexp.FindString(string(message[:msgLen]))

		logFile, ok := r.routerMap[clientAddr.String()]
		if !ok {
			r.log.WithFields(logrus.Fields{
				"address": clientAddr.String(),
				"server":  "unknown",
			}).Debugf(cleanMsg)
			continue
		}
		r.log.WithFields(logrus.Fields{
			"server": logFile.Name(),
			"state":  logFile.State(),
		}).Debugf(cleanMsg)
		logFile.Channel() <- cleanMsg
	}
}

func MakeRouterMap(hosts []config.Client, apiKey string, conn *mongo.Client, log *logrus.Logger) map[string]*server.LogFile {
	serverMap := make(map[string]*server.LogFile)
	client := &http.Client{Timeout: 10 * time.Second}
	for _, h := range hosts {
		s := server.NewLogFile(log, conn, h.Domain, h.Server)
		r := requests.NewRequester(apiKey, client)
		gi := stats.NewMatchData(h.Domain, h.Server)
		go server.StartWorker(log, s, r, gi)
		serverMap[h.Address] = s
		log.Infof("Started worker for %s#%d with host %s", h.Domain, h.Server, h.Address)
	}
	return serverMap
}
