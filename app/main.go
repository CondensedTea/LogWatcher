package main

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/logger"
	"LogWatcher/pkg/router"
	"context"
	"log"
)

var (
	LogLevel   = "debug"
	ConfigPath = "config.yaml"
)

func main() {
	ctx := context.Background()

	l, err := logger.NewLogger(LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logrus logger: %s", err)
	}

	l.Infof("Launching LogWatcher, log level is %s", LogLevel)

	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		l.Fatalf("Failed to load config: %s", err)
	}
	r, err := router.NewRouter(ctx, cfg, l)
	if err != nil {
		l.Fatalf("Failed to create Router: %s", err)
	}
	r.Listen()
}
