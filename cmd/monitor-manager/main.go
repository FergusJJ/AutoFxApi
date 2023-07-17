package main

import (
	monitormanager "api/internal/monitor-manager"
	"api/internal/storage"
	"api/pkg/shutdown"
	"log"
	"os"
)

//need to change testStoragePositions to makle sure that positions aren't mixed between different pools

//this should be ran in place of a single monitor
//I think maybe should connect to API
//then request should be sent to specific API route
//{"name":"monitor-url-name","type":"ct-markets","option":"start"}
//{"name":"monitor-url-name","type":"ct-markets","option":"stop"}
//on start call momnitor, run it, monitor should pipe terminal output to this program
//will also be in charge of restarting monitors that fail

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()

	cleanup, err := start()
	defer cleanup()
	if err != nil {
		log.Println(err)
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
	manager := monitormanager.NewManager()
	err = manager.Manage(client)
	//maybe cleanup will close individualmonitors as well
	return cleanup, err
}
