package main

import (
	"fmt"
	"log"
	"os"

	"api/config"
	"api/internal/http/middleware"
	"api/internal/http/router"
	"api/internal/storage/postgres"
	pgdb "api/internal/storage/postgres"
	cache "api/internal/storage/redis"
	"api/pkg/shutdown"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app         *fiber.App
	RedisClient *cache.RedisClientWithContext
	PGStore     pgdb.PGManager
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
	server.app = fiber.New(cfg)

	err := middleware.UseMiddlewares(server.app)
	if err != nil {
		return nil, err
	}

	client, cleanup, err := cache.RedisInitialise()
	if err != nil {
		return cleanup, err
	}
	server.RedisClient = client

	store, err := postgres.NewPostgresStore()
	if err != nil {
		return cleanup, err
	}
	cleanup, err = store.Init(cleanup)
	if err != nil {
		return cleanup, err
	}
	server.PGStore = store

	err = router.SetupRoutes(server.app, server.RedisClient, server.PGStore)
	if err != nil {
		return cleanup, err
	}
	return cleanup, nil
}

/*

psql -d copyfx_db -U user_default
 docker exec -it api-postgres-1 "bash"
*/
