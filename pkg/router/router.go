package router

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/mongo"
	"LogWatcher/pkg/requests"
	"LogWatcher/pkg/server"
	sm "LogWatcher/pkg/stateMachine"
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

type AddressTable map[string]*sm.StateMachine

type Router struct {
	address      *net.UDPAddr
	addressTable AddressTable
	log          *logrus.Logger
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
	r := requests.NewClient(cfg.Server.APIKey, client, log)

	addressTable := MakeAddressTable(cfg.Clients, log, mongoClient, r)
	return &Router{
		address:      udpAddr,
		addressTable: addressTable,
		log:          log,
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

		stateMachine, ok := r.addressTable[clientAddr.String()]
		if !ok {
			r.log.WithFields(logrus.Fields{
				"address": clientAddr.String(),
				"server":  "unknown",
			}).Debugf(cleanMsg)
			continue
		}
		r.log.WithFields(logrus.Fields{
			"server": stateMachine.File.Name(),
			"state":  stateMachine.State,
		}).Debugf(cleanMsg)
		stateMachine.Channel <- cleanMsg
	}
}

func MakeAddressTable(hosts []config.Client, log *logrus.Logger, inserter mongo.Inserter, uploader requests.LogUploader) AddressTable {
	addressTable := make(AddressTable)
	for _, host := range hosts {
		file := server.NewLogFile(host)
		match := stats.NewMatch(host)
		stateMachine := sm.NewStateMachine(log, file, uploader, match, inserter)
		go stateMachine.StartWorker()
		addressTable[host.Address] = stateMachine
		log.Infof("Started worker for %s#%d with host %s", host.Domain, host.Server, host.Address)
	}
	return addressTable
}
