package router

import (
	handler "api/internal/http/handlers"
	"api/internal/storage"
	wsHandler "api/internal/ws/handlers"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

var positions = make(chan string)

func SetupRoutes(app *fiber.App, redisClient *storage.RedisClientWithContext) error {
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
					_, ok := handler.WsPools[positionUpdate.Pool]
					if !ok {
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
	handleConfigureMonitorWrapper := func(c *fiber.Ctx) error {
		err := handler.HandleConfigureMonitorWrapper(c, redisClient)
		if err != nil {
			return err
		}
		return nil
	}

	app.Get("/whop/validate", handler.HandleWhopValidate)

	app.Get("/ws/monitor", websocket.New(handleWsMonitorWrapper))

	monitor := internal.Group("/monitor")

	monitor.Post("/configure-monitor", handleConfigureMonitorWrapper)

	app.Use(func(c *fiber.Ctx) error {
		c.SendStatus(404)
		return c.Next()
	})
	return nil
}
