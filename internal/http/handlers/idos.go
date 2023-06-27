package handler

import (
	"log"

	"github.com/gofiber/websocket/v2"
)

type Client struct {
	Ts          int
	WsConn      *websocket.Conn
	Pool        *Pool
	Id          string
	Overwritten bool
}

func (c *Client) Read() {
	defer func() {
		c.Pool.Unregister <- c
		c.WsConn.Close()
	}()
	for {
		messageType, p, err := c.WsConn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		c.Pool.Broadcast <- string(p)
		log.Println("message of type", messageType)
	}
}

type NewClient struct {
	Ts     int
	WsConn *websocket.Conn
	Id     string
	Pool   *Pool
}

var WsPool = &Pool{
	Unregister: make(chan *Client),
	WsClients:  make(map[string]*Client),
	Broadcast:  make(chan string),
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
	Unregister chan *Client
	WsClients  map[string]*Client //WsClients
	Broadcast  chan string        //message to send to all clients
}

func (pool *Pool) Start() {
	for {
		select {
		case client := <-pool.Unregister:
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
