package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/RecoBattle/cmd/config"
	"github.com/RecoBattle/internal/logger"
	"github.com/sirupsen/logrus"
)

func main() {

	logger.SetUpLogger()

	cfg := config.NewConfig()

	parseFlags(cfg)

	cnf, err := cfg.GetConfig(cfg.ConfigASR)
	if err != nil {
		logrus.Fatalf("cnf is not set. Error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)

	defer cancel()
}
