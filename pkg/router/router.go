package router

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/mongo"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	"LogWatcher/pkg/stats"
	"context"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

const timeout = 10

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

type Router struct {
	address   *net.UDPAddr
	routerMap map[string]*server.LogFile
	log       *logrus.Logger
}

func NewRouter(ctx context.Context, cfg *config.Config, log *logrus.Logger) (*Router, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		return nil, err
	}

	mongoClient, err := mongo.NewMongo(ctx, cfg.Server.DSN, cfg.Server.MongoDatabase, cfg.Server.MongoCollection)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: timeout * time.Second}
	r := requests.NewRequester(cfg.Server.APIKey, client)

	m := MakeRouterMap(cfg.Clients, log, mongoClient, r)
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

func MakeRouterMap(hosts []config.Client, log *logrus.Logger, i mongo.Inserter, r requests.LogProcessor) map[string]*server.LogFile {
	serverMap := make(map[string]*server.LogFile)
	for _, h := range hosts {
		lf := server.NewLogFile(log, h.Domain, h.Server)
		md := stats.NewMatchData(h.Domain, h.Server)
		go server.StartWorker(log, lf, r, md, i)
		serverMap[h.Address] = lf
		log.Infof("Started worker for %s#%d with host %s", h.Domain, h.Server, h.Address)
	}
	return serverMap
}
