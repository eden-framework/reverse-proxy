package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
)

const (
	PacketBytesPrefix = "reverse-proxy-rpc"
)

type UnpackFunc bufio.SplitFunc

var prefixLength = len(PacketBytesPrefix)

func InternalUnpack(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// 校验数据有效位必须大于前缀+4个字节（数据包长度）
	if atEOF || len(data) <= prefixLength+4 {
		return
	}

	// 读取实际数据长度
	var length uint32
	err = binary.Read(bytes.NewReader(data[prefixLength:prefixLength+4]), binary.BigEndian, &length)
	if err != nil {
		return
	}

	availableLength := int(length) + prefixLength + 4
	if availableLength <= len(data) {
		return availableLength, data[prefixLength+4 : availableLength], nil
	}
	return
}
