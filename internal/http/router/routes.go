package router

import (
	handler "api/internal/http/handlers"
	"api/internal/storage/postgres"
	cache "api/internal/storage/redis"
	wsHandler "api/internal/ws/handlers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var positions = make(chan string)

func SetupRoutes(app *fiber.App, redisClient *cache.RedisClientWithContext, PGManager postgres.PGManager) error {
	go func() {
		staleCheckInterval := time.Second * 10
		ticker := time.NewTicker(staleCheckInterval)
		checkPositionUpdateInterval := time.Millisecond * 500
		checkPosition := time.NewTicker(checkPositionUpdateInterval)

		for {
			select {
			case <-ticker.C:
				staleIds := []string{}
				for k, v := range handler.ActiveClients {
					//if 10 seconds have passed and an Id does not have an associated connection, assume something went wrong.
					if v.Ts < int(time.Now().UnixMilli())-10000 && v.WsConn == nil {
						staleIds = append(staleIds, k)
					}
				}
				for _, id := range staleIds {
					for _, pool := range handler.ActiveClients[id].Pool {
						pool.Unregister <- handler.ActiveClients[id]
					}
					delete(handler.ActiveClients, id)

					log.Println("removed id:", id)
					//need to unsubscribe from pools too

				}
				log.Println("currentIds len: ", len(handler.ActiveClients))
			case <-checkPosition.C:
				positionUpdate, err := redisClient.PopPositionUpdate()
				if err != nil {
					log.Fatal(err)
				}

				if positionUpdate != nil {
					log.Print("new position update")
					log.Print(positionUpdate)
					_, ok := handler.WsPools[positionUpdate.Pool]
					if !ok {
						log.Print("pool not found, no connected sessions to broadcast")
						break
					}
					if len(handler.WsPools[positionUpdate.Pool].WsClients) > 0 {
						handler.WsPools[positionUpdate.Pool].Broadcast <- positionUpdate

					}

				}

			}
		}
	}()

	internal := app.Group("/internal")
	handleWsMonitorWrapper := func(c *websocket.Conn) {
		wsHandler.HandleWsMonitor(c)
	}
	handleConfigureMonitor := func(c *fiber.Ctx) error {
		err := handler.HandleConfigureMonitorWrapper(c, redisClient)
		if err != nil {
			return err
		}
		return nil
	}

	/*ACCOUNT ROUTES START*/
	//Account create is probably redundant, will just lock behind bearer
	handleCreateAccount := func(c *fiber.Ctx) error {
		err := handler.HandleCreateAccountWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleDeleteAccount := func(c *fiber.Ctx) error {
		err := handler.HandleDeleteAccountWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleUpdateAccount := func(c *fiber.Ctx) error {
		err := handler.HandleUpdateAccountWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleGetAccount := func(c *fiber.Ctx) error {
		err := handler.HandleGetAccountWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	/*ACCOUNT ROUTES END*/
	/*AUTH ROUTES START*/
	handleAuthAccount := func(c *fiber.Ctx) error {
		err := handler.HandleAuthAccountWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleRefreshToken := func(c *fiber.Ctx) error {
		err := handler.HandlerRefreshTokenWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	/*AUTH ROUTES END*/

	/*WHOP ROUTES START*/
	app.Get("/whop/validate", func(c *fiber.Ctx) error {
		err := handler.HandleWhopValidateWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	})
	/*WHOP ROUTES START*/

	/*USER ROUTES START*/
	handleGetAllUserPosition := func(c *fiber.Ctx) error {
		err := handler.HandleGetAllUserPositionWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleDeleteUserPosition := func(c *fiber.Ctx) error {
		err := handler.HandleDeleteUserPositionWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	handleNewUserPosition := func(c *fiber.Ctx) error {
		err := handler.HandleCreateUserPositionWrapper(c, PGManager)
		if err != nil {
			return err
		}
		return nil
	}
	/*USER ROUTES END*/

	app.Get("/ws/monitor", websocket.New(handleWsMonitorWrapper))

	monitor := internal.Group("/monitor")
	monitor.Post("/configure-monitor", handleConfigureMonitor)

	api := app.Group("/api")
	account := api.Group("/account")
	account.Post("/create", handleCreateAccount)
	account.Post("/delete", handleDeleteAccount)
	account.Post("/get", handleGetAccount)
	account.Post("/update", handleUpdateAccount)

	user := api.Group("/user")
	user.Get("/position/all", handleGetAllUserPosition)
	user.Post("/position/new", handleNewUserPosition)
	user.Post("/position/delete", handleDeleteUserPosition)

	//These routes need seperate grouping
	api.Get("/auth/new", handleAuthAccount)
	api.Post("/auth/refresh", handleRefreshToken)

	app.Use(func(c *fiber.Ctx) error {
		c.SendStatus(404)
		return c.Next()
	})
	return nil
}
