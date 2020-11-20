package codec

import (
	"bytes"
	"encoding/binary"
)

type PackFunc func(payload []byte) (packet []byte, err error)

func InternalPack(payload []byte) (packet []byte, err error) {
	packetLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(packetLengthBytes, uint32(len(payload)))

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(PacketBytesPrefix)
	buf.Write(packetLengthBytes)
	buf.Write(payload)

	return buf.Bytes(), nil
}
