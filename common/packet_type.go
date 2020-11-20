package common

//go:generate eden generate enum --type-name=PacketType
// api:enum
type PacketType uint8

func (v *PacketType) UnmarshalBinary(data []byte) error {
	if len(data) > 0 {
		*v = PacketType(data[0])
	}
	return nil
}

func (v PacketType) MarshalBinary() (data []byte, err error) {
	data = append(data, uint8(v))
	return
}

//
const (
	PACKET_TYPE_UNKNOWN        PacketType = iota
	PACKET_TYPE__HANDSHAKE                // handshake
	PACKET_TYPE__HANDSHAKE_ACK            // ack for handshake
	PACKET_TYPE__REGISTER                 // register
	PACKET_TYPE__REGISTER_ACK             // ack for register
)
