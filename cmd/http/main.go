package main

import (
	"fmt"
	"log"
	"os"

	"api/config"
	"api/internal/http/middleware"
	"api/internal/http/router"
	"api/internal/monitor"
	"api/internal/storage"
	"api/pkg/shutdown"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app            *fiber.App
	CurrentMessage chan ([]byte)
	RedisClient    *storage.RedisClientWithContext
}

var server = &Server{}

func main() {
	var exitCode int
	defer func() {
		os.Exit(exitCode)
	}()
	cfg := fiber.Config{
		AppName:       "Pollo API",
		CaseSensitive: true,
		Prefork:       false,
	}

	port, err := config.Config("PORT")
	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}
	cleanup, err := start(port, cfg)
	defer cleanup()
	if err != nil {
		fmt.Printf("error: %v", err)
		exitCode = 1
		return
	}
	shutdown.Gracefully()
}

func start(port string, cfg fiber.Config) (func(), error) {

	cleanup, err := server.buildServer(cfg)
	if err != nil {
		return nil, err
	}
	go func() {
		server.app.Listen(":" + port)
	}()

	return func() {
		log.Println("running cleanup...")
		cleanup()
		server.app.Shutdown()
	}, nil
}

func (server *Server) buildServer(cfg fiber.Config) (func(), error) {
	formattedPositions := make(chan struct{}, 1)
	server.app = fiber.New(cfg)
	server.CurrentMessage = make(chan []byte)
	err := middleware.UseMiddlewares(server.app)
	if err != nil {
		return nil, err
	}

	err = router.SetupRoutes(server.app, formattedPositions)
	if err != nil {
		return nil, err
	}

	client, cleanup, err := storage.RedisInitialise()
	if err != nil {
		return cleanup, err
	}
	server.RedisClient = client

	monitorSess, err := monitor.Initialise()
	if err != nil {
		log.Panicln(err)
	}

	monitor.Start(monitorSess, server.RedisClient, formattedPositions)
	go func() {
		for {
			select {
			case <-formattedPositions:
				log.Printf("got positon signal")

			}
		}
	}()
	return cleanup, nil
}

/*



 */
