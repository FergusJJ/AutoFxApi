package handlers

import (
	handler "api/internal/http/handlers"
	"log"
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
	_, ok := handler.WsPool.WsClients[incomingId] //handler.WsClients[incomingId]
	if !ok {
		log.Printf("id %s does not exist", incomingId)
		return
	}
	if handler.WsPool.WsClients[incomingId].WsConn != nil {
		//close old connection, this removes incomingID from map though currently
		cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "connection already registered with same ID, close previous connection before attempting to connect")
		if err := c.WriteMessage(websocket.CloseMessage, cm); err != nil {
			log.Println(err)
		}
	}

	handler.WsPool.WsClients[incomingId].WsConn = c

	defer func(id string) {
		unregister <- c //to stop ticker
		handler.WsPool.Unregister <- handler.WsPool.WsClients[id]
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
