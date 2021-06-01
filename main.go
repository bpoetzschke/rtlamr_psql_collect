package main

import (
	"os"
	"strings"

	"github.com/bpoetzschke/rtlamr_psql_collect/database"
	"github.com/bpoetzschke/rtlamr_psql_collect/repositories"
	"github.com/bpoetzschke/rtlamr_psql_collect/rtlamrclient"
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
		return
	}

	defer func() {
		closeErr := db.Close()
		if closeErr != nil {
			log.Errorf("Failed to close db connection: %s", closeErr)
		}
	}()

	repo := repositories.NewRTLAMRRepo(db)
	rtlAmr, err := rtlamrclient.New(repo)
	if err != nil {
		log.Errorf("Error while initializing rtl amr: %s", err)
		return
	}

	err = rtlAmr.Run()
	if err != nil {
		log.Errorf("Error while running rtl amr: %s", err)
	}
}
