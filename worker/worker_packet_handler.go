package worker

import (
	"bufio"
	"github.com/eden-framework/context"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"net"
)

func (w *Worker) handleMasterHandshakeAck(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	logrus.Debugf("handleMasterHandshakeAck packet received from: %s", conn.RemoteAddr())

	payload, err := w.r.MarshalBinary()
	if err != nil {
		logrus.Error("handleMasterHandshakeAck router.MarshalBinary() err: %v", err)
		return
	}
	packet := &common.Packet{
		Type:    common.PACKET_TYPE__REGISTER,
		Length:  uint32(len(payload)),
		Payload: payload,
	}

	writer := bufio.NewWriter(conn)
	_ = w.writePacket(writer, packet)
	err = writer.Flush()
	if err != nil {
		logrus.Errorf("handleMasterHandshakeAck err: %v", err)
	}
}

func (w *Worker) handleMasterRegisterAck(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	logrus.Debugf("handleMasterRegisterAck packet received from: %s", conn.RemoteAddr())

	r := NewRouter()
	err := r.UnmarshalBinary(p.Payload)
	if err != nil {
		logrus.Error("handleMasterRegisterAck router.UnmarshalBinary() err: %v", err)
		return
	}

	for port := range r.Routes {
		if route, ok := w.r.Routes[port]; ok {
			go route.Start(ctx, conn)
		}
	}
}
