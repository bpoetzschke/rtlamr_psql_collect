package main

import (
	"os"
	"strings"

	"github.com/bpoetzschke/rtlamr_psql_collect/database"
	log "github.com/sirupsen/logrus"
)

func setupLogger() {
	fullTimestampFormatter := log.TextFormatter{}
	fullTimestampFormatter.FullTimestamp = true
	log.SetFormatter(&fullTimestampFormatter)

	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {
	setupLogger()
	db, err := database.Init()
	if err != nil {
		log.Errorf("Failed to connect to database: %s", err)
		os.Exit(1)
	}

	defer db.Close()
}
