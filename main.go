package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bpoetzschke/rtlamr_psql_collect/database"
	"github.com/bpoetzschke/rtlamr_psql_collect/repositories"
	"github.com/bpoetzschke/rtlamr_psql_collect/rtlamrclient"
	"github.com/bpoetzschke/rtlamr_psql_collect/rtltcp"
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

func setupSignalHandlers(cancelFunc context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Kill, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case sig := <-sigChan:
				log.Infof("Received signal %s, cancel context.", sig.String())
				cancelFunc()
				return
			}
		}
	}()

}

func main() {
	setupLogger()

	ctx := context.Background()
	cancelCtx, cancelFunc := context.WithCancel(ctx)

	setupSignalHandlers(cancelFunc)

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

	rtlTCP := rtltcp.NewClient()
	go func() {
		err := rtlTCP.Run(cancelCtx)
		if err != nil {
			log.Errorf("Error while running rtl_tcp command: %s", err)
			cancelFunc()
		}
	}()

	log.Info("Start rtlamr")

	repo := repositories.NewRTLAMRRepo(db)
	rtlAmr, err := rtlamrclient.New(repo)
	if err != nil {
		log.Errorf("Error while initializing rtl amr: %s", err)
		return
	}

	err = rtlAmr.Run(cancelCtx)
	if err != nil {
		log.Errorf("Error while running rtl amr: %s", err)
	}
}
