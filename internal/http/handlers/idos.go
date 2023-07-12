package handler

import (
	"api/pkg/ctrader"
	"log"

	"github.com/gofiber/websocket/v2"
)

var WsPools = map[string]*Pool{
	"7venWwvj": {
		Id:         "7venWwvj",
		Unregister: make(chan *Client),
		WsClients:  make(map[string]*Client),
		Broadcast:  make(chan *ctrader.CtraderMonitorMessage),
	},
	"pool1": {
		Id:         "pool1",
		Unregister: make(chan *Client),
		WsClients:  make(map[string]*Client),
		Broadcast:  make(chan *ctrader.CtraderMonitorMessage),
	},
}

var ActiveClients = map[string]*Client{}

type Client struct {
	Ts     int
	WsConn *websocket.Conn
	Pool   []*Pool
	Id     string
}

type invalidRequestResponse struct {
	ResponseCode int    `json:"responseCode"`
	Message      string `json:"message"`
}

type validLicenseKeyResponse struct {
	ResponseCode int    `json:"responseCode"`
	Cid          string `json:"cid"`
}

type Pool struct {
	Id         string
	Unregister chan *Client
	WsClients  map[string]*Client                  //WsClients
	Broadcast  chan *ctrader.CtraderMonitorMessage //message to send to all clients
}

// pool.unreg is not being hit
func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Unregister:
			log.Println("here")
			delete(pool.WsClients, client.Id)
			client.WsConn.Close()

		case message := <-pool.Broadcast:
			for _, client := range pool.WsClients {
				if err := client.WsConn.WriteJSON(message); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}
}
