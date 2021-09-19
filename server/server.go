package main

import (
	"log"
	"net"
	"strings"
)

const configPath = "config.yaml"

func makeAddressMap(hosts []Client, apiKey string) map[string]*LogFile {
	logsDict := make(map[string]*LogFile)
	for _, h := range hosts {
		ch := make(chan string)
		lf := &LogFile{
			server:  h.Server,
			region:  h.Region,
			ip:      h.Address,
			state:   Pregame,
			channel: ch,
			apiKey:  apiKey,
		}
		go lf.StartWorker()
		logsDict[h.Address] = lf
		log.Printf("Started worker for %s#%d with address %s", lf.region, lf.server, lf.ip)
	}
	return logsDict
}

func main() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	m := makeAddressMap(cfg.Clients, cfg.Server.APIKey)

	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		log.Fatalf("Failed to parse UDP address: %s", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP listner: %s", err)
	}
	defer conn.Close()

	log.Printf("LogWatcher is listening on %s", udpAddr.String())
	for {
		message := make([]byte, 1024)
		msgLen, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Fatalf("Failed to read from UDP: %s", err)
		}

		cleanMsg := strings.TrimSpace(string(message[:msgLen]))

		clientHost := clientAddr.String()

		lf, ok := m[clientHost]
		if !ok {
			log.Printf("[Unknown message]: [%s] %s", clientHost, cleanMsg)
			continue
		}
		log.Printf("[%s][state:%d] %s", clientAddr.String(), lf.state, cleanMsg)
		lf.channel <- cleanMsg
	}
}
