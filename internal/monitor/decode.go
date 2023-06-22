package monitor

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"unicode/utf8"

	"github.com/mitchellh/mapstructure"
)

func (MessageBuf *MessageBuf) DecodeInitial() WsMessage[bytesPayload] {
	var initialDecode WsMessage[bytesPayload]
	messageProcessor_ := &messageProcessor{
		MessageBuf: MessageBuf,
		type_:      0,
		position:   0,
		path:       "",
	}
	initialDecode = messageProcessor_.parseInitialPayload(MessageBuf.MessageType)
	return initialDecode
}
func (MessageBuf *MessageBuf) DecodeSpecific(payloadType int) (interface{}, error) {
	var specificStruct interface{}
	messageProcessor_ := &messageProcessor{
		MessageBuf: MessageBuf,
		type_:      0,
		position:   0,
		path:       "",
	}
	decodedPayload := messageProcessor_.decodeToSpecificPayload(MessageBuf.MessageType, payloadType, "")
	switch payloadType {
	case 4315:
		specificStruct = ProtoJMGetSharingTraderRes{}
		if err := mapstructure.Decode(decodedPayload, &specificStruct); err != nil {
			return nil, err
		}
	case 4259:
		specificStruct = ProtoJMTraderPositionListRes{}
		if err := mapstructure.Decode(decodedPayload, &specificStruct); err != nil {
			return nil, err
		}

	default:
		log.Fatalf("uncaught %d", payloadType)
	}
	return specificStruct, nil
}

func (messageProcessor *messageProcessor) parseInitialPayload(msgTypeFields map[int][]string) (decodedWsMessage WsMessage[bytesPayload]) {
	for messageProcessor.position < len(messageProcessor.MessageBuf.Arr) {

		byteType := messageProcessor.I_()
		// msgTypePosition := byteType >> 3
		msgTypePosition := (byteType >> 3) //they don't 0-index their arrays
		currPos := messageProcessor.position
		messageProcessor.type_ = 7 & byteType
		_, ok := msgTypeFields[msgTypePosition]
		if ok {
			subArrayPosition := msgTypeFields[msgTypePosition][0] //the first index of msgTypeFields[msgTypePosition]
			subKey := msgTypeFields[msgTypePosition][1]           //should be the name of the key
			subValueType := msgTypeFields[msgTypePosition][2]     //the typing of the value associated with the key
			_ = ""                                                //I think this is just for msgTypes with arrays? Is only initialised if there is an extra entru in msgTypeFields
			if len(msgTypeFields[msgTypePosition+1]) > 4 {
				_ = msgTypeFields[msgTypePosition][4] //additionalParam
			}
			var keyIndexStr string
			if messageProcessor.path != "" {
				keyIndexStr = fmt.Sprintf("%s.", messageProcessor.path)
			}
			//parse types out of subkey and subarrayPosition, to just get the values
			_, subArrPositionVal := splitTypeVal(subArrayPosition)
			_, subKeyVal := splitTypeVal(subKey)
			keyIndexStr = fmt.Sprintf("%s%s[%s]", keyIndexStr, subKeyVal, subArrPositionVal)
			decodedWsMessage = DecodeInitial(messageProcessor, subKey, decodedWsMessage, keyIndexStr, subValueType)
		}
		if messageProcessor.position == currPos {
			log.Fatal("block not implemented")
		}
	}
	return decodedWsMessage
}

func (messageProcessor *messageProcessor) decodeToSpecificPayload(msgTypeFields map[int][]string, msgType int, pathIndex string, optionalUpperBound ...int) (decodedMsg map[string]interface{}) {
	decodedMsg = make(map[string]interface{})
	tmpMap := make(map[string][]interface{})
	upperBound := len(messageProcessor.MessageBuf.Arr)
	if len(optionalUpperBound) > 0 {
		upperBound = optionalUpperBound[0]
	}

	for messageProcessor.position < upperBound {
		messageProcessor.path = pathIndex
		byteType := messageProcessor.I_()
		// msgTypePosition := byteType >> 3
		msgTypePosition := (byteType >> 3) //they dont use an array, they use an object with keys starting from 1, checking for the key each time
		currPos := messageProcessor.position
		messageProcessor.type_ = 7 & byteType
		_, ok := msgTypeFields[msgTypePosition]
		if ok {
			subArrayPosition := msgTypeFields[msgTypePosition][0] //the first index of msgTypeFields[msgTypePosition]
			subKey := msgTypeFields[msgTypePosition][1]           //should be the name of the key
			subValueType := msgTypeFields[msgTypePosition][2]     //the typing of the value associated with the key //will be ["repeated-simple", f()] for those types of payloads
			_ = ""                                                //I think this is just for msgTypes with arrays? Is only initialised if there is an extra entru in msgTypeFields
			if len(msgTypeFields[msgTypePosition]) > 4 {
				_ = msgTypeFields[msgTypePosition][4]
			}
			var keyIndexStr string
			if pathIndex != "" {
				keyIndexStr = fmt.Sprintf("%s.", pathIndex)
			}
			//parse types out of subkey and subarrayPosition, to just get the values
			_, subArrPositionVal := splitTypeVal(subArrayPosition)
			_, subKeyVal := splitTypeVal(subKey)
			keyIndexStr = fmt.Sprintf("%s%s[%s]", keyIndexStr, subKeyVal, subArrPositionVal)
			key, val := decodePayloadBytes(messageProcessor, subKey, keyIndexStr, subValueType) //, msgType)
			if key == "out" {
				key = parseField(keyIndexStr)
				tmpMap[key] = append(tmpMap[key], val)
			} else {
				decodedMsg[key] = val
			}
		}
		if messageProcessor.position == currPos {
			log.Fatal("block not implemented")
		}
	}
	//need to check whether tmpMap has any keys
	for k, v := range tmpMap {
		if len(tmpMap[k]) > 0 {
			decodedMsg[k] = v
		}
	}
	return decodedMsg
}

func (messageProcessor *messageProcessor) R_(combinedPos int) string {
	if combinedPos-messageProcessor.position >= 50 {
		subarr := messageProcessor.MessageBuf.Arr[messageProcessor.position:combinedPos]
		runeVal, _ := utf8.DecodeRune(subarr)
		return string(runeVal)
	} else {
		//R = combinedPos
		return messageProcessor.E_(combinedPos)
	}

}

func DecodeInitial(messageProcessor *messageProcessor, subKey string, resultStruct WsMessage[bytesPayload], keyIndexStr, subValueType string) WsMessage[bytesPayload] {
	//will just have switch inside of here instead of checking for function, array, etc.
	//need to parse value from messageProcessor, add to resultStruct, on next iteration, same result struct will be passed through
	var decodedVal interface{}
	_, parsedSubValueType := splitTypeVal(subValueType)
	_, structField := splitTypeVal(subKey)

	switch parsedSubValueType {
	case "string":
		tmpVal := messageProcessor.I_() + messageProcessor.position
		decodedVal = messageProcessor.R_(tmpVal)
		messageProcessor.position = tmpVal
	case "uint32":
		decodedVal = messageProcessor.I_()
	case "bytes":
		tmpVal := messageProcessor.I_() + messageProcessor.position
		decodedVal = messageProcessor.MessageBuf.Arr[messageProcessor.position:tmpVal]
		messageProcessor.position = tmpVal

	}

	switch structField {
	case "clientMsgId":
		resultStruct.ClientMsgId = decodedVal.(string)
	case "payloadType":
		resultStruct.PayloadType = decodedVal.(int)
	case "payload":
		resultStruct.Payload.Bytes = decodedVal.([]byte)
	default:
		log.Fatal()
	}
	return resultStruct

}

func decodePayloadBytes(messageProcessor *messageProcessor, subKey, keyIndexStr, subValueType string) (string, interface{}) {
	//will just have switch inside of here instead of checking for function, array, etc.
	//need to parse value from messageProcessor, add to resultStruct, on next iteration, same result struct will be passed through
	var decodedVal interface{}
	valueType, parsedSubValueType := splitTypeVal(subValueType)
	_, structField := splitTypeVal(subKey)
	switch valueType {
	case "string":
		switch parsedSubValueType {
		case "string":
			//need to cast to string
			tmpVal := messageProcessor.I_() + messageProcessor.position
			decodedVal = messageProcessor.R_(tmpVal)
			messageProcessor.position = tmpVal

		case "float":
			decodedValStr := messageProcessor.j(true, 23, 4)
			messageProcessor.position += 4
			decodedVal, _ = strconv.ParseFloat(decodedValStr, 64)
			//need to parseFloat
		case "double":
			decodedValStr := messageProcessor.j(true, 52, 8)
			messageProcessor.position += 8
			decodedValTmp, err := strconv.ParseFloat(decodedValStr, 64)
			if err != nil {
				log.Fatal(err)
			}
			// decodedValStr = strconv.FormatFloat(decodedValTmp, 'f', -1, 64)
			// decodedValTmp, err = strconv.ParseFloat(decodedValStr, 64)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// decodedVal = math.Round(decodedValTmp*1e8) / 1e8
			decodedVal = decodedValTmp
			// decodedVal.(floa

		case "bool":
			// var decodedVal bool
			tmpDecodedVal := messageProcessor.I_()
			if tmpDecodedVal != 0 {
				decodedVal = true
			} else {
				decodedVal = false
			}

		case "enum":
			decodedVal = messageProcessor.I_()

		case "uint32":
			//need to parseInt
			decodedVal = messageProcessor.I_()

		case "uint64":
			v1, v2 := messageProcessor.S_()
			decodedVal = messageProcessor.N_(v1, v2, false)

		case "int32":
			decodedVal = 0 | messageProcessor.I_()

			//call P_ to normalise or whatever
			//then set the value in the resultStruct
			// resultStruct.Set(decodedValue)

		case "int64":
			v1, v2 := messageProcessor.S_()
			decodedVal = messageProcessor.N_(v1, v2, true)

		case "sint32":
			tmpVal := messageProcessor.I_()
			decodedVal = int(uint32(tmpVal)>>1) ^ -(1 & tmpVal) | 0

		case "sint64":
			v1, v2 := messageProcessor.S_()
			tmpVal := -(1 & v1)
			arg1 := int(uint32(((int(uint32(v1)>>1) | v2<<31) ^ tmpVal)) >> 0)
			arg2 := int(uint32((int(uint32(v2)>>1) ^ tmpVal)) >> 0)
			decodedVal = messageProcessor.N_(arg1, arg2, true)

		case "fixed32":
			decodedVal = messageProcessor.D_()

		case "fixed64":

			log.Fatalf("haven't implemented fixed64:\n%v", messageProcessor.MessageBuf)
			decodedVal = messageProcessor.D_() //+ messageProcessor.D_()*4294967296

		case "sfixed32":
			decodedVal = messageProcessor.C_()

		case "sfixed64":
			log.Fatalf("haven't implemented sfixed64:\n%v", messageProcessor.MessageBuf)
			// don't know value of k
			decodedVal = messageProcessor.D_() // + messageProcessor.C_()*4294967296

		case "bytes":
			tmpVal := messageProcessor.I_() + messageProcessor.position
			decodedVal = messageProcessor.MessageBuf.Arr[messageProcessor.position:tmpVal]
			messageProcessor.position = tmpVal

		}
	case "array":
		//going to assume repeated-simple

		subValueType = fmt.Sprintf("function:%s", parsedSubValueType)
		structField, decodedVal = decodePayloadBytes(messageProcessor, "string:out", keyIndexStr, subValueType) //, msgType)

	case "function":
		msgType := functionMessageId[parsedSubValueType]
		decodedVal = messageProcessor.decodeToSpecificPayload(SliceFromMessageType(msgType), msgType, keyIndexStr, messageProcessor.I_()+messageProcessor.position)

	default:
		log.Fatalf("default case hit: %s", valueType)
	}
	return structField, decodedVal
}

func (messageProcessor *messageProcessor) I_() int {
	decodedByte := 0 // R
	xorVal := 0      // E
	// i := 0
	for i := 0; i < 4; i++ {
		xorVal = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
		messageProcessor.position++
		decodedByte = int(uint32(decodedByte|(127&xorVal)<<(7*i)) >> 0)
		if xorVal < 128 {
			return decodedByte
		}
	}
	xorVal = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
	messageProcessor.position++
	decodedByte = int(uint32(decodedByte|(15&xorVal)<<28) >> 0)
	if xorVal < 128 {
		return decodedByte
	}
	messageProcessor.position += 5
	if messageProcessor.position > len(messageProcessor.MessageBuf.Arr) {
		log.Fatalf("payload position out of bounds")
	}
	return decodedByte
}

// E_
func (messageProcessor *messageProcessor) E_(combinedPos int) string {
	result := ""
	var S rune
	var N = 1
	tempPos := messageProcessor.position // D
	for tempPos < combinedPos {
		T := rune(messageProcessor.MessageBuf.Arr[tempPos])
		if T > 239 {
			N = 4
		} else if T > 223 {
			N = 3
		} else if T > 191 {
			N = 2
		} else {
			N = 1
		}
		if tempPos+N > combinedPos {
			break
		}
		switch N {
		case 1:
			if T < 128 {
				S = T
			}
		case 2:
			O := rune(messageProcessor.MessageBuf.Arr[tempPos+1])
			if (192 & O) == 128 {
				S = rune((31&T)<<6 | (63 & O))
				if S <= 127 {
					S = O
				}
			}
		case 3:
			O := rune(messageProcessor.MessageBuf.Arr[tempPos+1])
			A := rune(messageProcessor.MessageBuf.Arr[tempPos+2])
			if (192&O) == 128 && (192&A) == 128 {
				S = (15&T)<<12 | (63&O)<<6 | (63 & A)
				if S <= 2047 || (S >= 55296 && S <= 57343) {
					S = rune(0)
				}
			}
		case 4:
			O := rune(messageProcessor.MessageBuf.Arr[tempPos+1])
			A := rune(messageProcessor.MessageBuf.Arr[tempPos+2])
			I := rune(messageProcessor.MessageBuf.Arr[tempPos+3])
			if (192&O) == 128 && (192&A) == 128 && (192&I) == 128 {
				S = rune((15&T)<<18 | (63&O)<<12 | (63&A)<<6 | (63 & I))
				if S <= 65535 || S >= 1114112 {
					S = rune(0)
				}
			}
		}
		if S == rune(0) {
			S = rune(65533)
			N = 1
		} else if S > 65535 {
			S -= 65536
			result += string(rune((uint32(S) >> 10 & 1023) | 55296))
			S = 56320 | (1023 & S)
		}
		result += string(S)
		tempPos += N
	}
	return result
}

func (messageProcessor *messageProcessor) j(R bool, T, O int) string {
	var L, e int
	var A, I int
	S := 8*O - T - 1
	N := (1 << S) - 1
	C := N >> 1
	D := -7

	if R {
		e = O - 1
		L = -1
	} else {
		e = 0
		L = 1
	}
	t := messageProcessor.MessageBuf.Arr[messageProcessor.position+e]

	e += L
	A = int(t & ((1 << -D) - 1))
	t >>= -D
	D += S
	for D > 0 {
		A = 256*A + int(messageProcessor.MessageBuf.Arr[messageProcessor.position+e])
		e += L
		D -= 8
	}

	I = A & ((1 << -D) - 1)
	A >>= -D
	D += T
	for D > 0 {
		I = 256*I + int(messageProcessor.MessageBuf.Arr[messageProcessor.position+e])
		e += L
		D -= 8
	}

	if A == 0 {
		// log.Fatal("check this condition in j")
		A = int(1 - C)
	} else {
		if A == int(N) {
			log.Fatal("return I ? NaN : 1 / 0 * (t ? -1 : 1);")
		}
		I += int(math.Pow(2, float64(T)))
		A -= int(C)
	}

	sign := 1.0
	if t != 0 {
		sign = -1.0
	}
	return fmt.Sprintf("%v", sign*math.Pow(2, float64(int(A)-T))*float64(I))

}

func (messageProcessor *messageProcessor) C_() string {
	//(_.buf[_.pos++] | _.buf[_.pos++] << 8 | _.buf[_.pos++] << 16) + (_.buf[_.pos++] << 24)
	v1 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+1])
	v2 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+2]) << 8
	v3 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+3]) << 16
	v4 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+4]) << 24
	messageProcessor.position += 4

	return fmt.Sprint(v1 | v2 | v3 | v4)
}

func (messageProcessor *messageProcessor) D_() int {
	//(_.buf[_.pos++] | _.buf[_.pos++] << 8 | _.buf[_.pos++] << 16) + 16777216 * _.buf[_.pos++]
	v1 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+1])
	v2 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+2]) << 8
	v3 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+3]) << 16
	v4 := int(messageProcessor.MessageBuf.Arr[messageProcessor.position+3]) * 16777216
	v5 := int(v1 | v2 | v3)
	messageProcessor.position += 4
	return v4 + v5
}

// returns an array in the js, will check whether that is important or not
func (messageProcessor *messageProcessor) S_() (int, int) {
	R, E, T := 0, 0, 0
	if len(messageProcessor.MessageBuf.Arr)-messageProcessor.position <= 4 {
		for i := 0; i < 3; i++ {
			E = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
			messageProcessor.position++
			R = int(uint32(R|(127&E)<<(7*i)) >> 0) //converts to uint
			if E < 128 {
				return R, T
			}
		}
		E = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
		R = int(uint32(R|(127&E)<<21) >> 0)
		messageProcessor.position++
		return R, T

	}
	for i := 0; i < 4; i++ {
		E = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
		messageProcessor.position++
		R = int(uint32(R|(127&E)<<(7*i)) >> 0) //converts to uint
		if E < 128 {
			return R, T
		}
	}
	E = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
	messageProcessor.position++
	R = int(uint32(R|(127&E)<<28) >> 0)
	T = int(uint32(T|(127&E)>>4) >> 0)
	if E < 128 {
		return R, T
	}
	for i := 0; i < 5; i++ {
		E = int(messageProcessor.MessageBuf.Arr[messageProcessor.position])
		messageProcessor.position++
		T = int(uint32(T|(127&E)<<(7*i+3)) >> 0)
		if E < 128 {
			return R, T
		}
	}
	log.Fatal("invalid variant encoding")
	return 0, 0
}

func (messageProcessor *messageProcessor) N_(val1, val2 int, reversed bool) int {
	if reversed && ((uint32(val2) >> 31) != 0) {
		R := uint32(1+(^val1)) >> 0
		T := uint32(^val2) >> 0
		if R == 0 {
			T = uint32(T+1) >> 0
		}
		return -(int(R) + 4294967296*int(T))
	}
	return val1 + 4294967296*val2
}
