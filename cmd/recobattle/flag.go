package main

import (
	"flag"
	"os"

	"github.com/RecoBattle/cmd/config"
)

func parseFlags(cfg *config.ConfigData) {

	flag.StringVar(&cfg.RunAddr, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "adress connect database")
	flag.StringVar(&cfg.ConfigASR, "c", "../config/config.toml", "path config")
	flag.StringVar(&cfg.PathFileStorage, "s", "", "path file storage")

	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDRESS"); envRunAddr != "" {
		cfg.RunAddr = envRunAddr
	}

	if envDatabaseDSN := os.Getenv("DATABASE_URI"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	if envConfigASR := os.Getenv("CONFIG_ASR"); envConfigASR != "" {
		cfg.ConfigASR = envConfigASR
	}

	if envPathFileStorage := os.Getenv("PATH_FILE_STORAGE"); envPathFileStorage != "" {
		cfg.PathFileStorage = envPathFileStorage
	}
}
