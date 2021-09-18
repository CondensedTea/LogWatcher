package main

import (
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const configPath = "config.yaml"

type State int

const (
	Pregame State = iota
	Game    State = iota
)

type Client struct {
	Server  int    `yaml:"ID"`
	Region  string `yaml:"Region"`
	Address string `yaml:"Address"`
}

type Config struct {
	Server struct {
		Host   string `yaml:"Host"`
		APIKey string `yaml:"APIKey"`
	} `yaml:"Server"`
	Clients []Client `yaml:"Clients"`
}

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

func makeAddressMap(hosts []Client, apiKey string) map[string]LogFile {
	logsDict := make(map[string]LogFile)
	for _, h := range hosts {
		ch := make(chan string)
		logsDict[h.Address] = LogFile{
			server:  h.Server,
			region:  h.Region,
			ip:      h.Address,
			state:   Pregame,
			channel: ch,
			apiKey:  apiKey,
		}
	}
	return logsDict
}

func main() {
	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	addrs := makeAddressMap(cfg.Clients, cfg.Server.APIKey)

	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host)
	if err != nil {
		log.Fatalf("Failed to parse UDP address: %s", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP listner: %s", err)
	}
	defer conn.Close()

	exit := make(chan bool)

	for k := range addrs {
		go addrs[k].StartWorker(exit)
		log.Printf("Started worker for %s#%d with address %s", addrs[k].region, addrs[k].server, addrs[k].ip)
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

	select {
	case <-exit:
		os.Exit(0)
	}
}
