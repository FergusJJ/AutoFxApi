package handlers

import (
	handler "api/internal/http/handlers"
	"api/pkg/ctrader"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/websocket/v2"
)

var register = make(chan *websocket.Conn)

// var positions = make(chan []byte)
var unregister = make(chan *websocket.Conn)
var heartbeat = make(chan *websocket.Conn)
var ack = make(chan *websocket.Conn)
var errResp = make(chan *websocket.Conn)

func HandleWsMonitor(c *websocket.Conn, positions chan [][]byte) {
	// It seems the we only need one SocketListener goroutine for the whole server.
	// If this is the case, the next line should be moved outside of this func.

	// need to add c.Query("id") to a map to make sure that only the ids that are returned can communicate

	incomingId := c.Query("id")
	_, ok := handler.WsClients[incomingId]
	if !ok {
		log.Printf("id %s does not exist", incomingId)
		return
	}
	handler.WsClients[incomingId].WsConn = c

	defer func(id string) {
		unregister <- c
		delete(handler.WsClients, id)
		c.Close()
	}(incomingId)
	go func() {
		for {
			select {
			case <-positions:
				log.Println("positions recieved")
			}
		}
	}()
	go heartbeatMonitor(c, unregister)
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

func heartbeatMonitor(c *websocket.Conn, unregister chan *websocket.Conn) {
	heartbeatInterval := time.Second * 30
	heartbeatWait := time.Second * 10
	ticker := time.NewTicker(heartbeatInterval)
	for {
		select {
		case <-ticker.C:
			err := c.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(heartbeatWait))
			if err != nil {

				return
			}
		case <-unregister:
			ticker.Stop()
			log.Println("unregistering heartbeat, conn closed")
			return

		}
	}
}

func sendMessages(c *websocket.Conn, unregister chan *websocket.Conn) {
	dummyEventInterval := time.Second * 10 // this would be an incoming message from the webhook
	ticker := time.NewTicker(dummyEventInterval)

	for {
		select {
		case <-ticker.C:

			// newMessage := ctrader.CtraderMonitorMessage{
			// 	SymbolID: 43,
			// 	// Message:  time.Now().String(),
			// }
			// err := c.WriteJSON(newMessage)
			// if err != nil {
			// 	log.Println("in ticker event dispatcher:", err)
			// 	return
			// }
			message := []byte{}
			messageJson := &ctrader.CtraderMonitorMessage{}
			err := json.Unmarshal(message, messageJson)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("here %+v", messageJson)
		case <-unregister:
			ticker.Stop()
			log.Println("unregistering heartbeat, conn closed")
			return
		}
	}
}
