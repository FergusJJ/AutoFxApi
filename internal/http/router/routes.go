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
				log.Println("checking stale ids")
				staleIds := []string{}
				for k, v := range handler.WsPool.WsClients {
					//if 10 seconds have passed and an Id does not have an associated connection, assume something went wrong.
					if v.Ts < int(time.Now().UnixMilli())-10000 && v.WsConn == nil {
						staleIds = append(staleIds, k)
					}
				}
				for _, id := range staleIds {
					delete(handler.WsPool.WsClients, id)
					log.Println("removed id:", id)
				}
				log.Println("currentIds len: ", len(handler.WsPool.WsClients))
			case <-checkPosition.C:
				positionUpdate, err := redisClient.PopPositionUpdate()
				if err != nil {
					log.Fatal(err)
				}
				if positionUpdate != nil {
					if len(handler.WsPool.WsClients) > 0 {
						// positions <- positionUpdate //need to push update to pool
						handler.WsPool.Broadcast <- positionUpdate
					}
					continue
				}

			}
		}
	}()
	go handler.WsPool.Start()

	handleWsMonitorWrapper := func(c *websocket.Conn) {
		wsHandler.HandleWsMonitor(c)

	}

	app.Get("/whop/validate", handler.HandleWhopValidate)

	app.Get("/ws/monitor", websocket.New(handleWsMonitorWrapper))

	app.Use(func(c *fiber.Ctx) error {
		c.SendStatus(404)
		return c.Next()
	})
	return nil
}
