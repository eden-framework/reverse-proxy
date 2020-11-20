package worker

import "bufio"

type HandleFunc func(payload []byte) (response []byte, err error)

type Handler struct {
	HandleFunc HandleFunc
	SplitFunc  bufio.SplitFunc
}
