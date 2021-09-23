package main

import (
	"os"

	"github.com/sirupsen/logrus"
)

const (
	configPath = "config.yaml"
	//apiHost    = "localhost:8081"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
}

func main() {
	cfg, err := LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Server instance: %s", err)
	}

	server.Listen()
}
