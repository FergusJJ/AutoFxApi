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
	//pool needs to be an option
	if len(os.Args) == 1 {
		log.Println("no args provided")
		exitCode = 1
		return
	}
	var Pool = os.Args[1] // "7venWwvj"

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
	monitorSess, err := monitor.Initialise(Pool)
	if err != nil {
		return cleanup, err
	}
	err = monitor.Start(monitorSess, client)
	if err != nil {
		return cleanup, err
	}
	return cleanup, nil
}

//need to close db connection in cleanup
