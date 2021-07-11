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

func setupSignalHandlers(ctx context.Context) context.Context {
	cancelCtx, cancelFunc := context.WithCancel(ctx)
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

	return cancelCtx
}

func main() {
	setupLogger()

	ctx := context.Background()
	ctx = setupSignalHandlers(ctx)

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
	startedChan := make(chan struct{}, 1)
	err = rtlTCP.Run(ctx, startedChan)
	if err != nil {
		log.Errorf("Error while starting rtl_tcp command: %s", err)
		os.Exit(1)
	}

	log.Info("Wait until rtl_tcp command started.")
	<-startedChan

	log.Info("rtl_tcp command started, starting rtlamr command.")

	repo := repositories.NewRTLAMRRepo(db)
	rtlAmr, err := rtlamrclient.New(repo)
	if err != nil {
		log.Errorf("Error while initializing rtl amr: %s", err)
		return
	}

	err = rtlAmr.Run(ctx)
	if err != nil {
		log.Errorf("Error while running rtl amr: %s", err)
	}
}
