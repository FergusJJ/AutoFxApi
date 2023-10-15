package handlers

import (
	handler "api/internal/http/handlers"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/gofiber/websocket/v2"
)

var unregister = make(chan *websocket.Conn)
var heartbeat = make(chan *websocket.Conn)
var ack = make(chan *websocket.Conn)
var errResp = make(chan *websocket.Conn)

func HandleWsMonitor(c *websocket.Conn) {
	// It seems the we only need one SocketListener goroutine for the whole server.
	// If this is the case, the next line should be moved outside of this func.

	// need to add c.Query("id") to a map to make sure that only the ids that are returned can communicate

	incomingId := c.Query("id")
	_, ok := handler.ActiveClients[incomingId]
	if !ok {
		log.Printf("id %s does not exist", incomingId)
		return
	}

	//old code that allowed for new pools to be spun up per client request
	// monitorPools := c.Headers("Pools") //should send a "," seperated list of the pools that the client wants to subscribe to, if none are sent then return
	// if monitorPools == "" {
	// 	//close conn, send message
	// 	log.Printf("id %s has not specified any pools", incomingId)
	// 	return
	// }
	// poolsSlice := strings.Split(monitorPools, ",")

	//only want predefined pools
	poolsSlice := []string{"7venWwvj"}

	for _, pool := range poolsSlice {
		pool = strings.TrimSpace(pool)
		_, ok := handler.WsPools[pool] //if pool doesn't exist, create pool
		if !ok {
			log.Print("pool not found, adding pool")
			newPool := handler.NewPool(pool) //pool created and added to WsPools
			go newPool.Start()
		} else {
			log.Printf("pool: %s exists, total pools: %d", pool, len(handler.WsPools))
		}
		if handler.WsPools[pool].WsClients[incomingId] != nil {
			cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure, fmt.Sprintf("connection to pool: %s already registered, close previous connection before attempting to reconnect", pool))
			if err := c.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Println(err)
				return
			}
		}
	}
	//could maybe have something where it will only reject pools that are already initialised
	//but will just make sure that all pools are closed first
	handler.ActiveClients[incomingId].WsConn = c
	for _, pool := range poolsSlice {
		pool = strings.TrimSpace(pool)
		handler.WsPools[pool].WsClients[incomingId] = handler.ActiveClients[incomingId]
		handler.ActiveClients[incomingId].Pool = append(handler.ActiveClients[incomingId].Pool, handler.WsPools[pool])
	}

	defer func(id string) {
		unregister <- c //to stop ticker
		for _, pool := range handler.ActiveClients[id].Pool {
			handler.WsPools[pool.Id].Unregister <- handler.ActiveClients[id]
		}
		handler.ActiveClients[id].WsConn.Conn.Close()
		delete(handler.ActiveClients, id)

	}(incomingId)

	go sendMessages(c, unregister)
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return
		}
		if messageType == websocket.TextMessage {
			log.Println("got textmessage:", string(message))
		} else {
			log.Println("received message of type:", messageType)
		}

	}
}

func sendMessages(c *websocket.Conn, unregister chan *websocket.Conn) {

	heartbeatInterval := time.Second * 30
	heartbeatWait := time.Second * 10
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	for {
		select {
		case <-heartbeatTicker.C:
			err := c.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(heartbeatWait))
			if err != nil {

				return
			}
		case <-unregister:
			heartbeatTicker.Stop()
			return
		}
	}

}
