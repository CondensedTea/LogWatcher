package main

import (
	"fmt"
	"net"
	"regexp"

	"github.com/sirupsen/logrus"
)

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

type Server struct {
	address    *net.UDPAddr
	addressMap map[string]*LogFile
}

func makeAddressMap(hosts []Client, apiKey string) map[string]*LogFile {
	logsDict := make(map[string]*LogFile)
	for _, h := range hosts {
		lf := &LogFile{
			Server:  h.Server,
			Domain:  h.Domain,
			IP:      h.Address,
			State:   Pregame,
			channel: make(chan string),
			apiKey:  apiKey,
		}
		go lf.StartWorker()
		logsDict[h.Address] = lf
		log.Infof("Started worker for %s#%d with host %s", lf.Domain, lf.Server, lf.IP)
	}
	return logsDict
}

func NewServer(cfg *Config) (*Server, error) {
	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		return nil, err
	}
	m := makeAddressMap(cfg.Clients, cfg.Server.APIKey)

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
	//defer s.conn.Close()
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
			"server": fmt.Sprintf("%s#%d", lf.Domain, lf.Server),
			"state":  lf.State.String(),
		}).Infof(cleanMsg)
		lf.channel <- cleanMsg
	}
}
