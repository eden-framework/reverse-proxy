package worker

import (
	"github.com/robotic-framework/reverse-proxy/codec"
)

type HandleFunc func(payload []byte) (response []byte, err error)

type Handler struct {
	HandleFunc HandleFunc
	PackFunc   codec.PackFunc
	UnpackFunc codec.UnpackFunc
}
