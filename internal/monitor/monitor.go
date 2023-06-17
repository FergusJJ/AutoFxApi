package monitor

import (
	"api/internal/storage"
	"log"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
)

func Initialise() (*MonitorSession, error) {
	var session = &MonitorSession{}
	conn, err := GetConn("wss://h30.p.ctrader.com/")
	if err != nil {
		return nil, err
	}
	session.Client = &WsClient{Conn: conn, CurrentMessage: make(chan []byte)}
	session.TraderLogin = -1
	return session, nil
}

func Start(session *MonitorSession, redisClient *storage.RedisClientWithContext) {
	go session.readPump()
	go session.writePump()
	session.monitor()

}

func (session *MonitorSession) monitor() (exitCode int) {
	exitCode = 0
	msgUUID := uuid.NewString()
	sharingCodePayloadVal := SharingCodePayload{SharingCode: "7venWwvj"}
	message := WsMessage[SharingCodePayload]{
		ClientMsgId: msgUUID,
		Payload:     sharingCodePayloadVal,
		PayloadType: 4314,
	}
	wsBody := Encode[SharingCodePayload](message, false)
	err := session.Client.Conn.WriteMessage(websocket.BinaryMessage, wsBody)
	if err != nil {
		log.Fatalf("write error: %+v", err)
	}
	for {
		messageType, message, err := session.Client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
				exitCode = 2
			}
			return exitCode
		}
		if messageType == websocket.TextMessage {
			session.Client.CurrentMessage <- message
		} else if messageType == websocket.BinaryMessage {
			session.Client.CurrentMessage <- message
		} else {
			log.Fatalf("got message of type %d", messageType)

		}
	}
}

func (session *MonitorSession) writePump() {

	pollPositionInterval := time.Second * 10
	ticker := time.NewTicker(pollPositionInterval)
	for {
		select {
		case <-ticker.C:
			//send message
			if session.TraderLogin == -1 {
				//still waiting for 4315 to be sent
				log.Println("waiting for TraderLogin")
				continue
			}
			log.Println("requesting open positions")

		}
	}
}

func (session *MonitorSession) readPump() {
	log.Println("listening for new messages")
	for {
		select {
		case message := <-session.Client.CurrentMessage:
			messageBuffer := &MessageBuf{
				MessageType: SliceFromMessageType(1),
				Arr:         message,
			}
			initialDecode := messageBuffer.DecodeInitial()

			messageBuffer.Arr = initialDecode.Payload.Bytes
			messageBuffer.MessageType = SliceFromMessageType(initialDecode.PayloadType)
			decodedMessage, err := messageBuffer.DecodeSpecific(initialDecode.PayloadType)
			if err != nil {
				log.Fatal(err)
			}
			switch initialDecode.PayloadType {
			case 4258:
				//send off to database, compare the new positons with ones in db
				//send any new positions to app as new position
				//send pid of newly closed position to app

			}
			log.Println(decodedMessage)
			//decide what to do with message based on the message type
			// _ = nil
		}
	}
}
