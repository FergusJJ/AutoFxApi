package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/fasthttp/websocket"
)

type Client struct {
	C              *websocket.Conn
	CurrentMessage chan []byte
}

func main() {

	// log.Println("doing stuff...")
	// time.Sleep(60 * time.Second)
	// log.Println("finished")
	mainThread()
}

func mainThread() {
	wsClient := &Client{}
	go wsClient.pollServer()
	//need to get the message from the channel
	for {
		select {
		case <-wsClient.CurrentMessage:
			fmt.Println("in here")
			//want to execute any messages that are sent from the server here
		default:
			continue
		}
	}
}

func (client *Client) pollServer() {
	addr := flag.String("addr", "localhost:8080", "http service address")
	u := url.URL{
		Scheme:   "ws",
		Host:     *addr,
		Path:     "/ws/monitor",
		RawQuery: "id=12",
	}

	log.Printf("connecting to %s\n", u.String())
	// Connect to the WebSocket server
	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Dial:", err)
	}
	client.C = conn
	if resp != nil {
		log.Println("Got response:", resp)
	}
	defer client.closeConn()
	client.listenConn()

}

func (client *Client) closeConn() {
	err := client.C.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		log.Println("Write close:", err)
		return
	}
	client.C.Close()
	log.Println("Connection closed")
}

func (client *Client) listenConn() {
	log.Println("started listening")
	for {
		_, message, err := client.C.ReadMessage()
		if err != nil {
			log.Println("client read error:", err)
			return
		}
		// log.Printf("Got message of type: %d,\nMessage: %s\n", messageType, string(message))
		client.CurrentMessage <- message
	}
}
