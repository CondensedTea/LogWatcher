package main

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/logger"
	"LogWatcher/pkg/router"
	"context"
	"flag"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	logLevel := flag.String("log", "debug", "Logging level")
	flag.Parse()
	ctx := context.Background()

	l, err := logger.NewLogger(*logLevel)
	if err != nil {
		log.Fatalf("Failed to create logrus logger: %s", err)
	}
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		l.Fatalf("Failed to parse config: %s", err)
	}
	conn, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Server.DSN))
	if err != nil {
		l.Fatalf("Failed to connect to mongodb: %s", err)
	}
	if err = conn.Ping(ctx, nil); err != nil {
		l.Warnf("Failed to ping mongodb: %s", err)
	}
	server, err := router.NewRouter(cfg, conn, l)
	if err != nil {
		l.Fatalf("Failed to create Router instance: %s", err)
	}
	server.Listen()
}
