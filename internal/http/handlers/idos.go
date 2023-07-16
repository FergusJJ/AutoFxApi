package handler

import (
	monitormanager "api/internal/monitor-manager"
	"api/pkg/ctrader"
	"log"

	"github.com/gofiber/websocket/v2"
)

/*
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
*/

var WsPools = map[string]*Pool{}
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

type monitorUpdateResponse struct {
	MonitorManagerMessage *monitormanager.MonitorManagerMessage `json:"monitorManagerMessage"`
	IsUpdate              bool                                  `json:"isUpdate"`
}

type Pool struct {
	Id         string
	Unregister chan *Client
	WsClients  map[string]*Client                  //WsClients
	Broadcast  chan *ctrader.CtraderMonitorMessage //message to send to all clients
}

func NewPool(name string) *Pool {
	newPool := &Pool{
		Id:         name,
		Unregister: make(chan *Client),
		WsClients:  make(map[string]*Client),
		Broadcast:  make(chan *ctrader.CtraderMonitorMessage),
	}
	WsPools[name] = newPool
	return newPool
}

// pool.unreg is not being hit
func (pool *Pool) Start() {
	defer func() {
		log.Println("deleting pool", pool.Id)
		delete(WsPools, pool.Id)
	}()
	for {
		select {
		case client := <-pool.Unregister:
			//remove client from pool
			log.Printf("len %s = %d", pool.Id, len(pool.WsClients))
			delete(pool.WsClients, client.Id)

			if len(pool.WsClients) == 0 {
				return
			}

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
