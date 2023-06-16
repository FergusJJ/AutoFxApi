package monitor

import (
	"log"
	"strconv"
)

type MessageBuf struct {
	Arr         []byte
	MessageType map[int][]string
}

func (MessageBuf *MessageBuf) Push(itemToPush []byte) {
	MessageBuf.Arr = append(MessageBuf.Arr, itemToPush...)
}

func (messsageBuf *MessageBuf) Concat(itemToPush []byte) {
	messsageBuf.Push(itemToPush)
}
func (messsageBuf *MessageBuf) Get() []byte {
	return messsageBuf.Arr
}

// these Encode functions should produce the same output as the functions within the "h_" object
// they are called within X_
func (MessageBuf *MessageBuf) StringEncoder(positionOfSubArray int, stringToEncode string) {
	//will change the underlying []byte so will just use receiver
	MessageBuf.BitShiftAndXor(positionOfSubArray, 2)
	MessageBuf.EncodeDataAsSlice(stringToEncode)
}

// K_
func (MessageBuf *MessageBuf) IntUintEncoder(positionOfSubArray int, stringToEncode string) {
	MessageBuf.BitShiftAndXor(positionOfSubArray, 0)
	encodeVal, err := strconv.Atoi(stringToEncode)
	if err != nil {
		log.Fatal(err)
	}
	MessageBuf.ModifyBuf(encodeVal, false)
}

func (MessageBuf *MessageBuf) BytesEncoder(positionOfSubArray int, stringToEncode string) {
	MessageBuf.BitShiftAndXor(positionOfSubArray, 2)
	byteSlice := []byte(stringToEncode)
	sliceLen := len(byteSlice)
	MessageBuf.ModifyBuf(sliceLen, false)
	MessageBuf.Push(byteSlice)
}

// used by all of the ${type}Encoder functions
// d_
func (MessageBuf *MessageBuf) BitShiftAndXor(lhsShift int, rhsXOR int) {
	MessageBuf.ModifyBuf(lhsShift<<3|rhsXOR, false)
}

// B_, this function has a default value for its third argument (set to false, but I haven't seen it used at all so will omit completely)
func (MessageBuf *MessageBuf) ModifyBuf(bitShiftedVal int, defaultFalseVal bool) {

	if bitShiftedVal > 268435455 || bitShiftedVal < 0 || defaultFalseVal {
		log.Fatalf("ModifyBuf: %d", bitShiftedVal)
		return //Q_ called here
	}
	unsignedBitShiftVal := uint32(bitShiftedVal)
	for unsignedBitShiftVal > 127 {
		valToPush := byte(127&unsignedBitShiftVal) | 128

		MessageBuf.Push([]byte{valToPush})
		unsignedBitShiftVal >>= 7
	}
	MessageBuf.Push([]byte{byte(unsignedBitShiftVal)})
}

// y_
// dataToPush is always converted to string before being added to the bytearray before being passed into y_
func (MessageBuf *MessageBuf) EncodeDataAsSlice(dataToPush string) {
	tempSlice := []byte{} //E
	// R := -1
	T := rune(-1)
	O := 0
	for ; O < len(dataToPush); O++ {
		// temp := int(dataToPush[O])

		r := []rune(dataToPush)[O]
		if r > 0xD7FF && r < 0xE000 {
			// character at index 0 is a surrogate pair, something to do with UTF-16
			if T == rune(-1) {
				if r > 0xDBFF || O+1 == len(dataToPush) {
					tempSlice = append(tempSlice, 239, 191, 189)
				} else {
					T = r
				}
				continue
			}
			if r < 56319 {
				tempSlice = append(tempSlice, 239, 191, 189)
				T = r
				continue
			}
			r = T - 55296<<10 | r - 56320 | 65536
			T = rune(-1)
		} else {
			if T != rune(-1) {
				tempSlice = append(tempSlice, 239, 191, 189)
				T = rune(-1)
			}
		}
		if r < 128 {
			//make sure there is no unexpected behaviour when converting types
			// if r > 255 {
			// 	log.Fatalf("run of value: %x exceeds the maximum value for byte", r)
			// } not needed lol
			tempSlice = append(tempSlice, byte(r))
		} else if r < 2048 {
			tempSlice = append(tempSlice, byte(r>>6|192)) //convert to proper range?
		} else if r < 65536 {
			tempSlice = append(tempSlice, byte(r>>12|224))
		} else {
			tempSlice = append(tempSlice, byte(r>>18|240), byte(r>>12&63|128), byte(r>>6&63|128), byte(63&r|128))
		}

	}
	MessageBuf.ModifyBuf(len(tempSlice), false)
	MessageBuf.Concat(tempSlice)
}

func (MessageBuf *MessageBuf) AddBytesToSlice(byteSlice []byte) {
	sliceLen := len(byteSlice)
	MessageBuf.ModifyBuf(sliceLen, false)
	MessageBuf.Push(byteSlice)

}
