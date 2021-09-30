package main

import (
	"context"
	"flag"
	"os"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	conn, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Server.DSN))
	if err != nil {
		log.Fatalf("Failed to connect to mongodb: %s", err)
	}
	server, err := NewServer(cfg, conn)
	if err != nil {
		log.Fatalf("Failed to create Server instance: %s", err)
	}

	server.Listen()
}
