package main

import (
	"api/internal/monitor"
	"api/internal/storage"
	"api/pkg/shutdown"
	"log"
	"os"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	session, err := monitor.Initialise()
	if err != nil {
		log.Fatalf("failed to start monitor: %v", err)
	}
	cleanup, err := start(session)
	defer cleanup()
	if err != nil {
		exitCode = 1
		return
	}
	shutdown.Gracefully()
}

func start(sess *monitor.MonitorSession) (func(), error) {

	client, cleanup, err := storage.RedisInitialise()
	if err != nil {
		return cleanup, err
	}
	go func() {
		monitorSess, err := monitor.Initialise()
		if err != nil {
			log.Panicln(err)
			return
		}
		monitor.Start(monitorSess, client)
	}()
	return cleanup, nil
}

//need to close db connection in cleanup
