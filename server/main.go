package main

import (
	"flag"
	"os"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
}

func main() {
	config := flag.String("config", "config.yaml", "Path to config file")
	flag.Parse()
	cfg, err := LoadConfig(*config)
	if err != nil {
		log.Fatalf("Failed to parse config: %s", err)
	}
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Server instance: %s", err)
	}

	server.Listen()
}
