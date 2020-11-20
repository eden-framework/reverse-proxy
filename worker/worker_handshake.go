package worker

import (
	"bufio"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"time"
)

func (w *Worker) handshake(writer *bufio.Writer) error {
	var err error
	for i := 0; i < w.RetryMaxTime; i++ {
		packet := &common.Packet{
			Type: common.PACKET_TYPE__HANDSHAKE,
		}
		err = w.writePacket(writer, packet)
		err = writer.Flush()
		if err != nil {
			logrus.Warningf("send packet err: %v", err)
			time.Sleep(w.RetryInterval)
		} else {
			logrus.Debugf("handshake sent")
			break
		}
	}
	return err
}
