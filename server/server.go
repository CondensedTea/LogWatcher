package main

import (
	"io/ioutil"
	"strings"

	//	"io/ioutil"
	"log"
	"net"
	//	"gopkg.in/yaml.v2"

	"golang.org/x/sync/errgroup"
	"gopkg.in/yaml.v2"
)

const configPath = "config.yaml"

func loadConfig(path string) (*Config, error) {
	var config Config
	yamlFile, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func makeAddressMap(hosts []Client) map[string]LogFile {
	logsDict := make(map[string]LogFile)
	for _, h := range hosts {
		ch := make(chan string)
		logsDict[h.Address] = LogFile{
			label:   h.Name,
			ip:      h.Address,
			state:   Pregame,
			channel: ch,
		}
	}
	return logsDict
}

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	addrs := makeAddressMap(cfg.Clients)

	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		log.Fatalf("Failed to parse UDP address: %s", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP listner: %s", err)
	}
	defer conn.Close()

	var g errgroup.Group

	for k := range addrs {
		g.Go(addrs[k].StartWorker)
		log.Printf("Started worker for %s with address %s", addrs[k].label, addrs[k].ip)
	}

	log.Printf("LogWatcher is listening on %s", udpAddr.String())

	go func() {
		for {
			message := make([]byte, 1024)
			msgLen, clientAddr, err := conn.ReadFromUDP(message)
			if err != nil {
				log.Fatalf("Failed to read from UDP: %s", err)
			}
			addressIP := clientAddr.IP.String()

			cleanMsg := strings.TrimSpace(string(message[:msgLen]))
			addrs[addressIP].channel <- cleanMsg
		}
	}()

	if err := g.Wait(); err != nil {
		log.Fatalf("Got error from worker: %s", err)
	}
}
