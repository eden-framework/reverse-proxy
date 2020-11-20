package common

import (
	"bytes"
	"encoding/binary"
)

type Packet struct {
	Type     PacketType
	Sequence uint32
	Length   uint32
	Payload  []byte
}

func (p *Packet) UnmarshalBinary(data []byte) error {
	reader := bytes.NewReader(data)
	typeByte, err := reader.ReadByte()
	if err != nil {
		return err
	}

	err = p.Type.UnmarshalBinary([]byte{typeByte})
	if err != nil {
		return err
	}

	intBytes := make([]byte, 4)
	_, err = reader.Read(intBytes)
	if err != nil {
		return err
	}
	p.Sequence = binary.BigEndian.Uint32(intBytes)

	_, err = reader.Read(intBytes)
	if err != nil {
		return err
	}
	p.Length = binary.BigEndian.Uint32(intBytes)

	if p.Length > 0 {
		payloadBytes := make([]byte, p.Length)
		_, err = reader.Read(payloadBytes)
		if err != nil {
			return err
		}
		p.Payload = payloadBytes[:]
	}

	return nil
}

func (p Packet) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer([]byte{})

	typeBytes, _ := p.Type.MarshalBinary()
	_, err = buf.Write(typeBytes)
	if err != nil {
		return
	}

	intBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(intBytes, p.Sequence)
	_, err = buf.Write(intBytes)
	if err != nil {
		return
	}

	binary.BigEndian.PutUint32(intBytes, p.Length)
	_, err = buf.Write(intBytes)
	if err != nil {
		return
	}

	_, err = buf.Write(p.Payload)
	if err != nil {
		return
	}

	data = buf.Bytes()
	return
}
