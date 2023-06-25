package monitor

import (
	"api/internal/storage"
	"api/pkg/ctrader"
	"encoding/json"
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
	session.Client = &WsClient{Conn: conn, CurrentMessage: []byte{}}
	session.TraderLogin = make(chan int)
	session.PlantID = make(chan string)
	session.Positions = make(chan []OpenPosition, 1)
	session.FormattedPositions = [][]byte{}
	return session, nil
}

func Start(session *MonitorSession, redisClient *storage.RedisClientWithContext, signalNewPositions chan struct{}) {

	go session.forwardPosititons(redisClient, signalNewPositions)
	go session.writePump()
	go session.monitor()
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
		_, message, err := session.Client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
				exitCode = 2
			}
			return exitCode
		}

		session.Client.CurrentMessage = message
		session.processMessage()
	}
}

func (session *MonitorSession) writePump() {
	var PlantId = ""
	var TraderLogin = -1
	pollPositionInterval := time.Millisecond * 1000
	ticker := time.NewTicker(pollPositionInterval)
	for {
		select {
		case <-ticker.C:
			if TraderLogin == -1 || PlantId == "" {
				continue
			}
			// log.Println("requesting open positions")

			msgUUID := uuid.NewString()
			positionRequestPayload := ProtoJMTraderPositionListReq{
				Cursor:      "",
				Limit:       1000,
				SharingCode: "7venWwvj",
				PlantId:     PlantId,
				TraderLogin: TraderLogin,
			}
			message := WsMessage[ProtoJMTraderPositionListReq]{
				ClientMsgId: msgUUID,
				Payload:     positionRequestPayload,
				PayloadType: 4258,
			}
			wsBody := Encode[ProtoJMTraderPositionListReq](message, false)
			err := session.Client.Conn.WriteMessage(websocket.BinaryMessage, wsBody)
			if err != nil {
				log.Fatalf("write error: %+v", err)
			}

		case PlantId = <-session.PlantID:
		case TraderLogin = <-session.TraderLogin:
		}
	}
}

func (session *MonitorSession) processMessage() {

	messageBuffer := &MessageBuf{
		MessageType: SliceFromMessageType(1),
		Arr:         session.Client.CurrentMessage,
	}
	initialDecode := messageBuffer.DecodeInitial()

	messageBuffer.Arr = initialDecode.Payload.Bytes

	switch initialDecode.PayloadType {
	case 4315:

		messageBuffer.MessageType = SliceFromMessageType(initialDecode.PayloadType)
		decodedMessage, err := messageBuffer.DecodeSpecific(initialDecode.PayloadType)
		if err != nil {
			log.Fatal(err)
		}

		message, ok := decodedMessage.(ProtoJMGetSharingTraderRes)
		if !ok {
			log.Fatal("couldn't cast message to ProtoJMGetSharingTraderRes")
		}
		session.TraderLogin <- message.TraderLogin
		session.PlantID <- message.PlantID
		// log.Println(4315)
	case 4259:

		messageBuffer.MessageType = SliceFromMessageType(initialDecode.PayloadType)
		decodedMessage, err := messageBuffer.DecodeSpecific(initialDecode.PayloadType)
		if err != nil {
			log.Fatal(err)
		}

		message, ok := decodedMessage.(ProtoJMTraderPositionListRes)
		if !ok {
			log.Fatal("couldn't cast message to ProtoJMTraderPositionListRes")
		}

		session.Positions <- message.Position

	default:
		//other messages that aren't needed, just skip

	}

}

func (session *MonitorSession) forwardPosititons(redisClient *storage.RedisClientWithContext, signalNewPositions chan struct{}) {

	var positionsName = "testStoragePositions"
	for {
		select {
		case <-session.Positions:
			positions := <-session.Positions
			// positions = []OpenPosition{}
			if len(positions) == 0 {
				log.Println("pos is 0 ")
				continue
			}
			positionMapping := positionsToPIDSlice(positions)
			pidsSlice := []string{}
			for k := range positionMapping {
				pidsSlice = append(pidsSlice, k)
			}
			closedPositions, openPositions, err := redisClient.ComparePositions(positionsName, pidsSlice)
			if err != nil {
				log.Fatal(err)
			}
			for _, pid := range closedPositions {
				direction := ""
				if positionMapping[pid].TradeSide == 2 {
					direction = "SELL"
				} else {
					direction = "BUY"
				}
				currentMessageStruct := ctrader.CtraderMonitorMessage{
					CopyPID:     pid,
					SymbolID:    positionMapping[pid].Symbol.SymbolID,
					Price:       positionMapping[pid].CurrentPrice, //send current price if position is closed
					Volume:      positionMapping[pid].Volume,
					Direction:   direction,
					MessageType: "CLOSE",
				}
				messageBytes, err := json.Marshal(currentMessageStruct)
				// _, err := json.Marshal(currentMessageStruct)
				if err != nil {
					log.Fatal(err)
				}
				session.FormattedPositions = append(session.FormattedPositions, messageBytes)

				// nextMessage <- messageBytes

			}
			for _, pid := range openPositions {
				direction := ""
				if positionMapping[pid].TradeSide == 2 {
					direction = "SELL"
				} else {
					direction = "BUY"
				}
				currentMessageStruct := ctrader.CtraderMonitorMessage{
					CopyPID:     pid,
					SymbolID:    positionMapping[pid].Symbol.SymbolID,
					Price:       positionMapping[pid].EntryPrice, //send entry price if position is opened
					Volume:      positionMapping[pid].Volume,
					Direction:   direction,
					MessageType: "OPEN",
				}
				messageBytes, err := json.Marshal(currentMessageStruct)
				// _, err := json.Marshal(currentMessageStruct)
				if err != nil {
					log.Fatal(err)
				}
				session.FormattedPositions = append(session.FormattedPositions, messageBytes)
				// nextMessage <- messageBytes
			}
			//signal that positions have been appended
			if len(session.FormattedPositions) > 0 {
				signalNewPositions <- struct{}{}
			}

		}
	}
}
