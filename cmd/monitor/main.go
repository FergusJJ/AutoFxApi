package main

import (
	"api/internal/monitor"
	"api/internal/storage"
	"api/pkg/shutdown"
	"log"
	"os"
)

func main() {
	var Pool = "7venWwvj"
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	cleanup, err := start(Pool)
	defer cleanup()
	if err != nil {
		log.Println(err)
		exitCode = 1
		return
	}
	shutdown.Gracefully()
}

func start(Pool string) (func(), error) {

	client, cleanup, err := storage.RedisInitialise()
	if err != nil {
		return cleanup, err
	}
	monitorSess, err := monitor.Initialise()
	if err != nil {
		return cleanup, err
	}
	err = monitor.Start(monitorSess, client, Pool)
	if err != nil {
		return cleanup, err
	}
	return cleanup, nil
}

//need to close db connection in cleanup
