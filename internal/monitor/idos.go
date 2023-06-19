package monitor

import (
	"log"

	"github.com/fasthttp/websocket"
)

type WsClient struct {
	Conn           *websocket.Conn
	CurrentMessage []byte
}

// Encode.go
type WsMessage[T ctPayload] struct {
	ClientMsgId string `json:"clientMsgId"`
	Payload     T      `json:"payload"`
	PayloadType int    `json:"payloadType"`
}

// outgoing messages before they are sent
type ctPayload interface {
	Get(string) interface{}
}

type SharingCodePayload struct {
	SharingCode string `json:"sharingCode"`
}

func (p SharingCodePayload) Get(key string) interface{} {
	switch key {
	case "sharingCode":
		return p.SharingCode
	default:
		return nil
	}
}

type bytesPayload struct {
	Bytes []byte `json:"bytes"`
}

func (p bytesPayload) Get(key string) interface{} {
	switch key {
	case "bytes":
		return p.Bytes
	default:
		return nil // log.Fatalf("key: %s does not exist in struct \"bytePayload\"", key)
	}
}

type ProtoJMGetSharingTraderRes struct {
	PlantID                     string `json:"plantId"`
	TraderLogin                 int    `json:"traderLogin"`
	TraderRegistrationTimestamp int    `json:"traderRegistrationTimestamp"`
	DepositCurrency             string `json:"depositCurrency"`
	Balance                     int    `json:"balance"`
	Deleted                     bool   `json:"deleted"`
	TotalMarginCalculationType  int    `json:"totalMarginCalculationType"`
	Nickname                    string `json:"nickname"`
	LeverageInCents             int    `json:"leverageInCents"`
	Live                        bool   `json:"live"`
	TraderAccountType           int    `json:"traderAccountType"`
	MoneyDigits                 int    `json:"moneyDigits"`
	DepositCurrencyDigits       int    `json:"depositCurrencyDigits"`
	Environment                 string `json:"environment"`
}

type genericRequest struct {
	PlantId     string `json:"PlantId"`
	SharingCode string `json:"sharingCode"`
	TraderLogin int    `json:"traderLogin"`
}

type ProtoJMTraderPositionListReq struct {
	Cursor      string `json:"cursor"`
	Limit       int    `json:"limit"`
	PlantId     string `json:"PlantId"`
	SharingCode string `json:"sharingCode"`
	TraderLogin int    `json:"traderLogin"`
}

func (p ProtoJMTraderPositionListReq) Get(key string) interface{} {
	switch key {
	case "cursor":
		return p.Cursor
	case "limit":
		return p.Limit
	case "plantId":
		return p.PlantId
	case "sharingCode":
		return p.SharingCode
	case "traderLogin":
		return p.TraderLogin
	default:
		return nil
	}
}

// 4259
type ProtoJMTraderPositionListRes struct {
	Position []OpenPosition `json:"position"`
}

type OpenPosition struct {
	PositionID                  int     `json:"positionId"`
	PositionStatus              int     `json:"positionStatus"`
	TradeSide                   int     `json:"tradeSide"`
	Symbol                      Symbol  `json:"symbol"`
	Volume                      int     `json:"volume"`
	EntryPrice                  float64 `json:"entryPrice"`
	OpenTimestamp               int     `json:"openTimestamp"`
	UtcLastUpdateTimestamp      int     `json:"utcLastUpdateTimestamp"`
	Commission                  int     `json:"commission"`
	Swap                        int     `json:"swap"`
	MarginRate                  float64 `json:"marginRate"`
	Profit                      int     `json:"profit"`
	ProfitInPips                float64 `json:"profitInPips"`
	CurrentPrice                float64 `json:"currentPrice"`
	Comment                     string  `json:"comment"`
	Channel                     string  `json:"channel"`
	MirroringCommission         int     `json:"mirroringCommission"`
	UsedMargin                  int     `json:"usedMargin"`
	IntroducingBrokerCommission int     `json:"introducingBrokerCommission"`
	MoneyDigits                 int     `json:"moneyDigits"`
	PnlConversionFee            int     `json:"pnlConversionFee"`
}
type Symbol struct {
	SymbolName  string `json:"symbolName"`
	Digits      int    `json:"digits"`
	PipPosition int    `json:"pipPosition"`
	SymbolID    int    `json:"symbolId"`
	Description string `json:"description"`
}

func SliceFromMessageType(MessageType int) map[int][]string {

	switch MessageType {
	//this is the format that all incoming messages are in and outgoing messages are sent in, message specific info kept in payload
	case 1:
		//x_ function
		return map[int][]string{
			1: {"int:1", "string:payloadType", "string:uint32", "int:1"},
			2: {"int:2", "string:payload", "string:bytes", "int:0"},
			3: {"int:3", "string:clientMsgId", "string:string", "int:0"},
		}

	case 4258:
		return map[int][]string{
			1: {"int:1", "string:payloadType", "string:enum", "int:0"},
			2: {"int:2", "string:traderLogin", "string:int64", "int:1"},
			3: {"int:3", "string:plantId", "string:string", "int:1"},
			4: {"int:4", "string:limit", "string:int32", "int:0"},
			5: {"int:5", "string:cursor", "string:string", "int:0"},
			6: {"int:6", "string:fromTimestamp", "string:int64", "int:0"},
			7: {"int:7", "string:toTimestamp", "string:int64", "int:0"},
			8: {"int:8", "string:sharingCode", "string:string", "int:0"}}

	case 4259:
		return map[int][]string{
			1: {"int:1", "string:payloadType", "string:enum", "int:0"},
			2: {"int:2", "string:position", "array:SE", "int:1"},
			3: {"int:3", "string:nextCursor", "string:string", "int:0"}}
		// return

	case 42590:
		return map[int][]string{
			1:  {"int:1", "string:positionId", "string:int64", "int:1"},             //
			2:  {"int:2", "string:positionStatus", "string:enum", "int:1"},          //
			3:  {"int:3", "string:tradeSide", "string:enum", "int:1"},               //
			4:  {"int:4", "string:symbol", "function:IE", "int:1"},                  //
			5:  {"int:5", "string:volume", "string:int64", "int:1"},                 //
			6:  {"int:6", "string:entryPrice", "string:double", "int:1"},            //
			7:  {"int:7", "string:openTimestamp", "string:int64", "int:1"},          //
			8:  {"int:8", "string:utcLastUpdateTimestamp", "string:int64", "int:1"}, //
			9:  {"int:9", "string:commission", "string:int64", "int:0"},             //
			10: {"int:10", "string:swap", "string:int64", "int:0"},                  //
			11: {"int:11", "string:closeTimestamp", "string:int64", "int:0"},        //
			12: {"int:12", "string:stopLossPrice", "string:double", "int:0"},        //
			13: {"int:13", "string:takeProfitPrice", "string:double", "int:0"},      //
			14: {"int:14", "string:marginRate", "string:double", "int:0"},           //shouldn't be 0, should be a decimal
			15: {"int:15", "string:profit", "string:int64", "int:0"},                //
			16: {"int:16", "string:profitInPips", "string:double", "int:0"},         //
			17: {"int:17", "string:currentPrice", "string:double", "int:0"},         //
			18: {"int:18", "string:comment", "string:string", "int:0"},              //
			19: {"int:19", "string:channel", "string:string", "int:0"},              //
			20: {"int:20", "string:label", "string:string", "int:0"},                //
			22: {"int:22", "string:mirroringCommission", "string:int64", "int:0"},
			23: {"int:23", "string:usedMargin", "string:uint64", "int:0"},
			24: {"int:24", "string:introducingBrokerCommission", "string:int64", "int:0"}, //
			25: {"int:25", "string:moneyDigits", "string:uint32", "int:0"},                //incorrect, when this is being decoded, the position is 3 ahead of where it should be i think
			26: {"int:26", "string:pnlConversionFee", "string:sint64", "int:0"},           //incorrect
			27: {"int:27", "string:netProfit", "string:int64", "int:0"}}                   //

	case 425900:
		return map[int][]string{
			1: {"int:1", "string:symbolName", "string:string", "int:1"},
			2: {"int:2", "string:digits", "string:int32", "int:1"},
			3: {"int:3", "string:pipPosition", "string:int32", "int:1"},
			4: {"int:4", "string:symbolId", "string:int64", "int:0"},
			5: {"int:5", "string:description", "string:string", "int:0"},
			6: {"int:6", "string:baseAssetName", "string:string", "int:0"},
			7: {"int:7", "string:baseAssetType", "string:enum", "int:0"},
			8: {"int:8", "string:assetClassDefaultLots", "string:bool", "int:0"}}

	case 4314: //initialRequest
		return map[int][]string{1: {"int:1", "string:payloadType", "string:enum", "int:0"}, 2: {"int:2", "string:sharingCode", "string:string", "int:1"}}

	case 4315:
		return map[int][]string{
			1:  {"int:1", "string:payloadType", "string:enum", "int:0"},
			2:  {"int:2", "string:plantId", "string:string", "int:1"},
			3:  {"int:3", "string:traderLogin", "string:int64", "int:1"},
			4:  {"int:4", "string:traderRegistrationTimestamp", "string:int64", "int:0"},
			5:  {"int:5", "string:depositCurrency", "string:string", "int:0"},
			6:  {"int:6", "string:balance", "string:int64", "int:0"},
			7:  {"int:7", "string:deleted", "string:bool", "int:0"},
			8:  {"int:8", "string:totalMarginCalculationType", "string:enum", "int:0"},
			9:  {"int:9", "string:nickname", "string:string", "int:0"},
			10: {"int:10", "string:leverageInCents", "string:int32", "int:0"},
			11: {"int:11", "string:live", "string:bool", "int:0"},
			12: {"int:12", "string:traderAccountType", "string:enum", "int:0"},
			13: {"int:13", "string:moneyDigits", "string:uint32", "int:0"},
			14: {"int:14", "string:depositCurrencyDigits", "string:uint32", "int:0"},
			15: {"int:15", "string:environment", "string:string", "int:0"},
		}

	default:
		log.Panicf("no messageTypeFunc defined for %d", MessageType)
	}
	return map[int][]string{}
}

// IE

var functionMessageId = map[string]int{
	"SE": 42590,
	"IE": 425900,
}

type messageProcessor struct {
	MessageBuf *MessageBuf
	position   int
	type_      int
	path       string //The path of the current value in the object that it will be store
}

type MonitorSession struct {
	Client      *WsClient
	TraderLogin chan (int)
	PlantID     chan (string)
	Positions   chan ([]OpenPosition) //use to send from monitor reader -> redis checker
}
