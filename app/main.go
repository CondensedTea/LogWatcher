package main

import (
	"LogWatcher/pkg/config"
	"LogWatcher/pkg/logger"
	"LogWatcher/pkg/router"
	"context"
	"log"
)

var ConfigPath = "config.yaml"

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig(ConfigPath)
	if err != nil {
		log.Fatalf("Failed to load config: %s", err)
	}

	l, err := logger.NewLogger(cfg.Server.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logrus logger: %s", err)
	}

	l.Infof("Launching LogWatcher, log level is %q", cfg.Server.LogLevel)

	r, err := router.NewRouter(ctx, cfg, l)
	if err != nil {
		l.Fatalf("Failed to create Router: %s", err)
	}
	r.Listen()
}
