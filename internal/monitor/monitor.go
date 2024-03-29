package monitor

import (
	cache "api/internal/storage/redis"
	"api/pkg/ctrader"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"syscall"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/google/uuid"
)

func Initialise(Pool string) (*MonitorSession, error) {
	log.Printf("initialising monitor %s\n", Pool)
	var session = &MonitorSession{}
	conn, err := GetConn("wss://h30.p.ctrader.com/")
	if err != nil {
		return nil, err
	}
	session.Client = &WsClient{Conn: conn, CurrentMessage: []byte{}}
	session.TraderLogin = make(chan int)
	session.PlantID = make(chan string)
	session.Pool = Pool
	return session, nil
}

func Start(session *MonitorSession, redisClient *cache.RedisClientWithContext) (err error) {
	unexpectedError := false
	go session.writePump()
	for !unexpectedError {
		err = session.monitor(redisClient)
		if !strings.Contains(err.Error(), "unexpected EOF") {
			unexpectedError = true
		}

	}

	return err
}

func (session *MonitorSession) monitor(redisClient *cache.RedisClientWithContext) (err error) {
	msgUUID := uuid.NewString()
	sharingCodePayloadVal := SharingCodePayload{SharingCode: session.Pool}
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
		positions, pass := session.processMessage()
		if !pass {
			continue
		}
		// if len(positions) == 0 {
		// 	continue
		// }

		err = session.forwardPosititons(session.Pool, redisClient, positions)
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
				SharingCode: session.Pool,
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

func (session *MonitorSession) processMessage() ([]OpenPosition, bool) {

	messageBuffer := &MessageBuf{
		MessageType: SliceFromMessageType(1),
		Arr:         session.Client.CurrentMessage,
	}
	initialDecode := messageBuffer.DecodeInitial()

	messageBuffer.Arr = initialDecode.Payload.Bytes
	switch initialDecode.PayloadType {
	case 50:
		messageBuffer.MessageType = SliceFromMessageType(initialDecode.PayloadType)
		decodedMessage, err := messageBuffer.DecodeSpecific(initialDecode.PayloadType)
		if err != nil {
			log.Fatal(err)
		}
		message, ok := decodedMessage.(ProtoErrorRes)
		if !ok {
			log.Fatal("couldn't cast message to ProtoErrorRes")
		}
		if message.ErrorCode == "CH_TRADER_ACCOUNT_NOT_FOUND" {
			log.Fatalf("monitor error: %s, for %s\n", message.Description, session.Pool)
		}
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
		// log.Printf("%+v", message)
		positions := []OpenPosition{}
		for _, pos := range message.Position {
			pos.Volume = pos.Volume / 100
			positions = append(positions, pos)
		}
		return positions, true
	default:
	}
	return []OpenPosition{}, false

}

func (session *MonitorSession) forwardPosititons(pool string, redisClient *cache.RedisClientWithContext, positions []OpenPosition) error {
	var positionChanges = []ctrader.CtraderMonitorMessage{}
	var positionsName = fmt.Sprintf("storage-positions-pool-%s", pool)
	//if len positions is 0, do still need to check whether positions have been closed
	// if len(positions) == 0 {
	// 	log.Println("pos is 0 ")
	// 	return nil
	// }

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
			Symbol:          positionMapping[pid].Symbol.SymbolName,
			Pool:            session.Pool,
			CopyPID:         pid,
			SymbolID:        positionMapping[pid].Symbol.SymbolID,
			Price:           positionMapping[pid].CurrentPrice, //send current price if position is closed
			Volume:          positionMapping[pid].Volume,
			Direction:       direction,
			OpenedTimestamp: positionMapping[pid].OpenTimestamp,
			MessageType:     "CLOSE",
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
			Symbol:          positionMapping[pid].Symbol.SymbolName,
			Pool:            session.Pool,
			CopyPID:         pid,
			SymbolID:        positionMapping[pid].Symbol.SymbolID,
			Price:           positionMapping[pid].EntryPrice, //send entry price if position is opened
			Volume:          positionMapping[pid].Volume,
			Direction:       direction,
			OpenedTimestamp: positionMapping[pid].OpenTimestamp,
			MessageType:     "OPEN",
		}
		positionChanges = append(positionChanges, currentMessageStruct)

	}
	//signal that positions have been appended
	//if position changes are old, don't send, will also avoid a bunch of position changes being sent on monitor startup
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
