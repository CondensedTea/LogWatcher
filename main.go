package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"gopkg.in/yaml.v2"
)

const configPath = "config.yaml"

func loadConfig() (*Config, error) {
	var config Config
	yamlFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(yamlFile, config); err != nil {
		return nil, err
	}
	return &config, nil
}

func getAddressMap(hosts []Client) map[string]LogFile {
	logsDict := make(map[string]LogFile)
	for _, h := range hosts {
		ch := make(chan string)
		logsDict[h.address] = LogFile{
			state:   Pregame,
			channel: ch,
		}
		logsDict[h.address].StartWorker()
	}
	return logsDict
}

//func spawnWorkers(clients []Client, logfiles map[string]LogFile) {
//	for _, v := range clients {}
//		go func() {
//
//		}()
//}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}

	addrs := getAddressMap(cfg.Clients)

	udpAddr, err := net.ResolveUDPAddr("udp4", cfg.Server.Host+":"+cfg.Server.Port)
	if err != nil {
		log.Fatalf("Failed to parse UDP address: %s", err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatalf("Failed to create UDP listner: %s", err)
	}
	defer conn.Close()

	log.Printf("Log Server listening on %s\n", udpAddr.String())

	for {
		message := make([]byte, 1024)
		rlen, clientAddr, err := conn.ReadFromUDP(message)
		if err != nil {
			log.Fatalf("Failed to read from UDP: %s", err)
		}
		cleanMsg := string(message[:rlen])
		cleanMsg = strings.TrimLeft(cleanMsg, "L ")
		cleanMsg = strings.TrimSpace(cleanMsg)

		fmt.Printf(clientAddr.String())

		addrs[clientAddr.String()].channel <- cleanMsg

		fmt.Printf("[%s] : %s\n", clientAddr.String(), message)
	}
}
