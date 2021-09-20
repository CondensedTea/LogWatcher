package main

import (
	"log"
)

const (
	configPath = "config.yaml"
	httpHost   = "0.0.0.0:8081"
)

func main() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Server instance: %s", err)
	}

	router := NewRouter(server)
	go router.Run(httpHost)

	server.Listen()
}
