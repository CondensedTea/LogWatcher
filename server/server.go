package main

import (
	"log"
	"net"
	"regexp"
)

const configPath = "config.yaml"

var logLineRegexp = regexp.MustCompile(`L \d{2}/\d{2}/\d{4} - \d{2}:\d{2}:\d{2}: .+`)

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

		cleanMsg := logLineRegexp.FindString(string(message[:msgLen]))

		lf, ok := m[clientAddr.String()]
		if !ok {
			log.Printf("[Unknown server]: [%s] %s", clientAddr.String(), cleanMsg)
			continue
		}
		log.Printf("[%s][state:%d] %s", clientAddr.String(), lf.state, cleanMsg)
		lf.channel <- cleanMsg
	}
}
