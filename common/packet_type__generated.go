package common

import (
	"bytes"
	"encoding"
	"errors"

	github_com_eden_framework_enumeration "github.com/eden-framework/enumeration"
)

var InvalidPacketType = errors.New("invalid PacketType")

func init() {
	github_com_eden_framework_enumeration.RegisterEnums("PacketType", map[string]string{
		"REGISTER_ACK":  "ack for register",
		"REGISTER":      "register",
		"HANDSHAKE_ACK": "ack for handshake",
		"HANDSHAKE":     "handshake",
	})
}

func ParsePacketTypeFromString(s string) (PacketType, error) {
	switch s {
	case "":
		return PACKET_TYPE_UNKNOWN, nil
	case "REGISTER_ACK":
		return PACKET_TYPE__REGISTER_ACK, nil
	case "REGISTER":
		return PACKET_TYPE__REGISTER, nil
	case "HANDSHAKE_ACK":
		return PACKET_TYPE__HANDSHAKE_ACK, nil
	case "HANDSHAKE":
		return PACKET_TYPE__HANDSHAKE, nil
	}
	return PACKET_TYPE_UNKNOWN, InvalidPacketType
}

func ParsePacketTypeFromLabelString(s string) (PacketType, error) {
	switch s {
	case "":
		return PACKET_TYPE_UNKNOWN, nil
	case "ack for register":
		return PACKET_TYPE__REGISTER_ACK, nil
	case "register":
		return PACKET_TYPE__REGISTER, nil
	case "ack for handshake":
		return PACKET_TYPE__HANDSHAKE_ACK, nil
	case "handshake":
		return PACKET_TYPE__HANDSHAKE, nil
	}
	return PACKET_TYPE_UNKNOWN, InvalidPacketType
}

func (PacketType) EnumType() string {
	return "PacketType"
}

func (PacketType) Enums() map[int][]string {
	return map[int][]string{
		int(PACKET_TYPE__REGISTER_ACK):  {"REGISTER_ACK", "ack for register"},
		int(PACKET_TYPE__REGISTER):      {"REGISTER", "register"},
		int(PACKET_TYPE__HANDSHAKE_ACK): {"HANDSHAKE_ACK", "ack for handshake"},
		int(PACKET_TYPE__HANDSHAKE):     {"HANDSHAKE", "handshake"},
	}
}

func (v PacketType) String() string {
	switch v {
	case PACKET_TYPE_UNKNOWN:
		return ""
	case PACKET_TYPE__REGISTER_ACK:
		return "REGISTER_ACK"
	case PACKET_TYPE__REGISTER:
		return "REGISTER"
	case PACKET_TYPE__HANDSHAKE_ACK:
		return "HANDSHAKE_ACK"
	case PACKET_TYPE__HANDSHAKE:
		return "HANDSHAKE"
	}
	return "UNKNOWN"
}

func (v PacketType) Label() string {
	switch v {
	case PACKET_TYPE_UNKNOWN:
		return ""
	case PACKET_TYPE__REGISTER_ACK:
		return "ack for register"
	case PACKET_TYPE__REGISTER:
		return "register"
	case PACKET_TYPE__HANDSHAKE_ACK:
		return "ack for handshake"
	case PACKET_TYPE__HANDSHAKE:
		return "handshake"
	}
	return "UNKNOWN"
}

var _ interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
} = (*PacketType)(nil)

func (v PacketType) MarshalText() ([]byte, error) {
	str := v.String()
	if str == "UNKNOWN" {
		return nil, InvalidPacketType
	}
	return []byte(str), nil
}

func (v *PacketType) UnmarshalText(data []byte) (err error) {
	*v, err = ParsePacketTypeFromString(string(bytes.ToUpper(data)))
	return
}
