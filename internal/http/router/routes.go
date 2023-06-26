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

func SetupRoutes(app *fiber.App, redisClient *storage.RedisClientWithContext) error {
	go func() {
		staleCheckInterval := time.Second * 10
		ticker := time.NewTicker(staleCheckInterval)
		for {
			select {
			case <-ticker.C:
				log.Println("checking stale ids")
				staleIds := []string{}
				for k, v := range handler.WsClients {
					//if 10 seconds have passed and an Id does not have an associated connection, assume something went wrong.
					if v.Ts < int(time.Now().UnixMilli())-10000 && v.WsConn == nil {
						staleIds = append(staleIds, k)
					}
				}
				for _, id := range staleIds {
					delete(handler.WsClients, id)
					log.Println("removed id:", id)
				}
			}
		}
	}()

	handleWsMonitorWrapper := func(c *websocket.Conn) {
		// go func () {

		// }()
		wsHandler.HandleWsMonitor(c, redisClient)

	}

	app.Get("/whop/validate", handler.HandleWhopValidate)

	app.Get("/ws/monitor", websocket.New(handleWsMonitorWrapper))

	app.Use(func(c *fiber.Ctx) error {
		c.SendStatus(404)
		return c.Next()
	})
	return nil
}
