package monitor

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func Encode[T ctPayload](message WsMessage[T], E bool) []byte {
	var bytePayloadType int = 1
	msgType := message.PayloadType //T
	var tmpMessage interface{}
	tmpMessage = message

	if E {
		log.Fatal("function: (Encode) is incomplete, need to handle first selection statement")
	} else {
		var encodedPayload []byte
		switch msgType {
		case 4314:
			encodedPayload = createPayloadFromTypeAndPayload[SharingCodePayload](tmpMessage.(WsMessage[SharingCodePayload]), msgType)
		case 4258:
			encodedPayload = createPayloadFromTypeAndPayload[ProtoJMTraderPositionListReq](tmpMessage.(WsMessage[ProtoJMTraderPositionListReq]), msgType)
		default:
			log.Fatalf("message.PayloadType %d not included", message.PayloadType)
		}
		// encodedPayload := createPayloadFromTypeAndPayload(Tval, message.Payload) //.get() is called on the MessageBuf class which returns the underlying array
		payloadBody := bytesPayload{
			Bytes: encodedPayload,
		}
		payload2 := WsMessage[bytesPayload]{
			PayloadType: msgType,
			Payload:     payloadBody,
			ClientMsgId: message.ClientMsgId,
		}

		encodedPayload = createPayloadFromTypeAndPayload[bytesPayload](payload2, bytePayloadType)

		return encodedPayload
	}
	return []byte{}
}

func createPayloadFromTypeAndPayload[T ctPayload](payload WsMessage[T], msgType int) []byte {
	newMessageBuf := &MessageBuf{Arr: []byte{}}
	newMessageBuf.Arr = lowercaseBUnderscore(newMessageBuf, payload, msgType)
	return newMessageBuf.Arr
}

func lowercaseBUnderscore[T ctPayload](arrayToPassTo *MessageBuf, payload WsMessage[T], msgType int) []byte {
	var E = 0
	var payloadValueKey = 1
	var payloadValueType = 2
	var _ = 3 // I
	var _ = 4 //S
	msgTypeArray := SliceFromMessageType(msgType)
	//If subArray[O] is a field in payload, then call EncodeDataType
	for _, subArray := range msgTypeArray {
		if len(subArray) == 0 {
			continue
		}
		//if the subarray is of length 5, will set "T" (T is what is iterated through) to an object i think
		//then would set I to T.value, as opposed to
		if len(subArray) == 5 {
			//set some variable T to payload[S] (this may be some message info? This will be in a msgType array so will get to it when i get to an array of length 5)
			log.Panicf("Slice: %+v\nhas length of 4, need to finish func\n", msgTypeArray)
		} else if len(subArray) == 4 {
			var fieldVal interface{}
			var dataToEncode = ""
			_, val := splitTypeVal(subArray[payloadValueKey])
			fieldVal = payload.Payload.Get(val)
			if fieldVal == nil {
				if msgType != 1 {
					continue
				}
				if val == "clientMsgId" {
					fieldVal = payload.ClientMsgId
				}
				if val == "payloadType" {
					fieldVal = payload.PayloadType
				}
				if val == "payload" {
					fieldVal = payload.Payload.Get("bytes")
				}
			}
			interfaceTypeSlice := strings.Split(subArray[payloadValueType], ":")
			if len(interfaceTypeSlice) == 1 {
				log.Fatalf("error: interfaceTypeSlice has length 1")
			}
			//The [type]:[value] encoding here is unnecessary i think, need to remove
			interfaceType := interfaceTypeSlice[1]
			switch interfaceType {
			case "string":
				dataToEncode = "string:" + fieldVal.(string)
				if dataToEncode == "string:do_not_encode" {
					continue
				}
			case "int64", "int32":
				dataToEncode = fmt.Sprintf("int:%d", fieldVal.(int))
			case "uint32":
				dataToEncode = fmt.Sprintf("int:%d", fieldVal.(int))
			case "bytes":
				byteArr := fieldVal.([]byte)
				dataToEncode = fmt.Sprintf("bytes:%s", string(byteArr))
			default:
				log.Fatalf("interfaceType: %s is not included in switch\n", interfaceType)
			}

			encodeDataType(arrayToPassTo, subArray[payloadValueType], subArray[E], dataToEncode)

		} else {
			log.Panicf("Slice: %+v\n not len 4 or 5, need to finish func\n", msgTypeArray)
		}
	}

	return arrayToPassTo.Arr
}

func encodeDataType(MessageBuf *MessageBuf, E string, R string, T string) {
	//need to parse E, R, T into their values and typing
	_, Etype := splitTypeVal(E)
	Rtype, Rvalue := splitTypeVal(R)
	//first checks the value of E
	//E typeof string, does not check the value itself as it could be "int32", "enum", etc. need a way to specify the value as the value is a func, string or array..
	//could E with some value, and check the value against a map that stores the value that it should be converted to/treated as
	//if its an array could just stringify the array originally then parse it?
	//if its a function it looks like the return value is an array anyways, so could just make a map of function names to their return values as they will be static anyways

	//reversed the order of conditional as type of E is always string
	if Etype == "function" { //E typof function

	} else if Etype == "array" { // E typeof array (some of the msgType arrays contain arrays within them)
		//check array format
		//may want to stringify all values then convert to array within here??

	} else {
		rInt, err := strconv.Atoi(Rvalue)
		if err != nil {
			log.Fatalf("Rvalue is not of type int, (%s , %s)\n%+v", Rtype, Rvalue, err)
		}
		switch Etype {
		case "string":
			//know that type of T will always be string as Etype tells value of T
			_, Tval := splitTypeVal(T)
			MessageBuf.StringEncoder(rInt, Tval)
		case "bool":
			_, Tval := splitTypeVal(T)
			if Tval == "true" {
				MessageBuf.IntUintEncoder(rInt, "1")
			} else {
				MessageBuf.IntUintEncoder(rInt, "0")
			}
		case "enum", "uint32", "uint64", "int32", "int64":
			_, Tval := splitTypeVal(T)
			MessageBuf.IntUintEncoder(rInt, Tval)
		case "bytes":
			_, Tval := splitTypeVal(T)
			MessageBuf.BytesEncoder(rInt, Tval)
			// log.Fatal([]byte(Tval))

		default:
			log.Fatalf("Etype: %s, not handled in switch block", Etype)
		}

	}

}
