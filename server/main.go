package main

import (
	"log"
)

const configPath = "config.yaml"

func main() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server instance: %s", err)
	}

	//router := NewRouter(server)
	//go router.Run()

	server.Listen()
}
