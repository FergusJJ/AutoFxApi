package main

import (
	"fmt"
	"log"
	"os"

	"api/config"
	"api/internal/http/middleware"
	"api/internal/http/router"
	"api/pkg/shutdown"

	"github.com/gofiber/fiber/v2"
)

type Server struct {
	app *fiber.App
}

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

	app, cleanup, err := buildServer(cfg)
	if err != nil {
		return nil, err
	}
	go func() {
		app.Listen(":" + port)
	}()
	//want to make sure that ActiveLicenseKeys & WsClients are synced, in case

	return func() {
		log.Println("running cleanup...")
		cleanup()
		app.Shutdown()
	}, nil
}

func buildServer(cfg fiber.Config) (*fiber.App, func(), error) {
	app := fiber.New(cfg)

	err := middleware.UseMiddlewares(app)
	if err != nil {
		return nil, nil, err
	}

	err = router.SetupRoutes(app)
	if err != nil {
		return nil, nil, err
	}

	return app, func() {
		//cleanup operations, i.e. close db connection
		log.Println("no cleanup required")
	}, nil
}

/*



 */
