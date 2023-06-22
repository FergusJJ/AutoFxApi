package main

import (
	"api/internal/monitor"
	"api/internal/storage"
	"api/pkg/ctrader"
	"api/pkg/shutdown"
	"encoding/json"
	"log"
	"os"
)

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	cleanup, err := start()
	defer cleanup()
	if err != nil {
		exitCode = 1
		return
	}
	shutdown.Gracefully()
}

func start() (func(), error) {

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
		var testChan = make(chan []byte)
		var sigChan = make(chan struct{})
		monitor.Start(monitorSess, client, sigChan)
		for {
			select {
			case data := <-testChan:
				var message = &ctrader.CtraderMonitorMessage{}
				json.Unmarshal(data, message)
				log.Println(message.CopyPID)
			}
		}
	}()
	return cleanup, nil
}

//need to close db connection in cleanup
