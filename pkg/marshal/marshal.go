package marshal

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
)

func MarshalUint32(data uint32) []byte {
	payload := make([]byte, 4)
	binary.BigEndian.PutUint32(payload, data)

	return payload
}

func UnmarshalUint32(data []byte) uint32 {
	return uint32(binary.BigEndian.Uint32(data[:4]))
}

func MarshalInt64(data int64) []byte {
	payload := make([]byte, 8)
	binary.BigEndian.PutUint64(payload, uint64(data))

	return payload
}

func UnmarshalInt64(data []byte) int64 {
	return int64(binary.BigEndian.Uint64(data))
}

func MarshalFloat64(data float64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, data)
	if err != nil {
		log.Println(err)
	}
	return buf.Bytes()
}

func UnmarshalFloat64(data []byte) float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(data))
}

func MarshalString(data string) []byte {
	payload := bytes.Join([][]byte{MarshalUint32(uint32(len(data))), []byte(data)}, []byte{})

	return payload
}

func UnmarshalString(data []byte) string {
	lenData := UnmarshalUint32(data[:4])

	return string(data[4 : 4+lenData])
}

func MarshalUint32Array(data []uint32) []byte {
	lenData := len(data)
	payload := make([]byte, 4+lenData*4)

	copy(payload[:4], MarshalUint32(uint32(lenData)))

	for i := 0; i < lenData; i++ {
		copy(payload[4+i*4:(i+1)*4], MarshalUint32(uint32(data[i])))
	}

	return payload
}
