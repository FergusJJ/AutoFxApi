package handlers

import (
	handler "api/internal/http/handlers"
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

	monitorPools := c.Headers("Pools") //should send a "," seperated list of the pools that the client wants to subscribe to, if none are sent then return
	if monitorPools == "" {
		log.Printf("id %s has not specified any pools", incomingId)
		return
	}

	//check that all pools exist before connecting
	poolsSlice := strings.Split(monitorPools, ",")
	for _, pool := range poolsSlice {
		_, ok := handler.WsPools[pool]
		if !ok {
			cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "cannot connect to a copy session that does not exist")
			if err := c.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Println(err)
			}
		}
		//
		// log.Println(handler.WsPools[pool].WsClients[incomingId])

		if handler.WsPools[pool].WsClients[incomingId] != nil {
			cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection already registered with same ID, close previous connection before attempting to connect to new copy sessions")
			if err := c.WriteMessage(websocket.CloseMessage, cm); err != nil {
				log.Println(err)
			}
		}
	}
	handler.ActiveClients[incomingId].WsConn = c

	for _, pool := range poolsSlice {
		handler.WsPools[pool].WsClients[incomingId] = handler.ActiveClients[incomingId]
	}

	defer func(id string) {
		unregister <- c //to stop ticker
		for _, pool := range handler.ActiveClients[id].Pool {
			pool.Unregister <- handler.ActiveClients[id]
		}
		log.Printf("Pool1 clients: %v", handler.WsPools["pool1"].WsClients)
		log.Printf("7venWwvj clients: %v", handler.WsPools["7venWwvj"].WsClients)
		delete(handler.ActiveClients, id)
		log.Printf("Active clients: %v", handler.ActiveClients)
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
			log.Println("unregistering heartbeat, conn closed")
			heartbeatTicker.Stop()

			return
		}
	}

}
