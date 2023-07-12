package monitor

import (
	"api/internal/storage"
	"api/pkg/ctrader"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
)

/*
Docker output:


2023-06-28 18:16:25 2023/06/28 17:16:25 sending 2 position changes
2023-06-28 21:28:42 2023/06/28 20:28:42 websocket: close 1006 (abnormal closure): unexpected EOF
2023-06-28 21:28:42 2023/06/28 20:28:42 Redis connection closed successfully
2023-06-28 21:28:43 exit status 1

2023-06-29 14:34:52 2023/06/29 13:34:52 write error: write tcp 172.18.0.4:41152->85.234.140.193:443: write: broken pipe
2023-06-29 14:34:52 2023/06/29 13:34:52 write tcp 172.18.0.4:41152->85.234.140.193:443: write: broken pipe
2023-06-29 14:34:52 2023/06/29 13:34:52 Redis connection closed successfully
2023-06-29 14:34:52 exit status 1

*/

func Initialise() (*MonitorSession, error) {
	var session = &MonitorSession{}
	conn, err := GetConn("wss://h30.p.ctrader.com/")
	if err != nil {
		return nil, err
	}
	session.Client = &WsClient{Conn: conn, CurrentMessage: []byte{}}
	session.TraderLogin = make(chan int)
	session.PlantID = make(chan string)
	return session, nil
}

func Start(session *MonitorSession, redisClient *storage.RedisClientWithContext, Pool string) (err error) {
	unexpectedError := false
	go session.writePump()
	for !unexpectedError {
		err = session.monitor(redisClient, Pool)
		if !strings.Contains(err.Error(), "unexpected EOF") {
			unexpectedError = true
		}

	}

	return err
}

func (session *MonitorSession) monitor(redisClient *storage.RedisClientWithContext, Pool string) (err error) {
	msgUUID := uuid.NewString()
	sharingCodePayloadVal := SharingCodePayload{SharingCode: "7venWwvj"}
	message := WsMessage[SharingCodePayload]{
		ClientMsgId: msgUUID,
		Payload:     sharingCodePayloadVal,
		PayloadType: 4314,
	}
	wsBody := Encode[SharingCodePayload](message, false)
	err = session.Client.Conn.WriteMessage(websocket.BinaryMessage, wsBody)
	if err != nil {
		if errors.Is(err, syscall.EPIPE) {
			log.Print("This is broken pipe error")
		}
		log.Println("write error:", err)
		return err
	}

	for {
		_, message, err := session.Client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Println("read error:", err)
			}
			return err
		}

		session.Client.CurrentMessage = message
		positions := session.processMessage()
		if len(positions) == 0 {
			continue
		}
		err = session.forwardPosititons(redisClient, positions, Pool)
		if err != nil {
			log.Println(err)
			return err
		}
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

func (session *MonitorSession) processMessage() []OpenPosition {

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
		return message.Position
	}
	return []OpenPosition{}

}

func (session *MonitorSession) forwardPosititons(redisClient *storage.RedisClientWithContext, positions []OpenPosition, Pool string) error {
	var positionChanges = []ctrader.CtraderMonitorMessage{}
	var positionsName = "testStoragePositions"
	if len(positions) == 0 {
		log.Println("pos is 0 ")
		return nil
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
			Pool:        Pool,
			CopyPID:     pid,
			SymbolID:    positionMapping[pid].Symbol.SymbolID,
			Price:       positionMapping[pid].CurrentPrice, //send current price if position is closed
			Volume:      positionMapping[pid].Volume,
			Direction:   direction,
			MessageType: "CLOSE",
		}
		positionChanges = append(positionChanges, currentMessageStruct)

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
		log.Printf("%+v", positionMapping[pid].Symbol)
		positionChanges = append(positionChanges, currentMessageStruct)

	}
	//signal that positions have been appended
	if len(positionChanges) > 0 {
		//send new positions to redis
		log.Printf("sending %d position changes", len(positionChanges))
		for _, pos := range positionChanges {
			jsonBytes, err := json.Marshal(pos)
			if err != nil {
				log.Fatal(err)
			}
			redisClient.PushPositionUpdate(jsonBytes)
		}
	}
	return nil

}
