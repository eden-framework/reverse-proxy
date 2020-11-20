package master

import (
	"bufio"
	"github.com/eden-framework/context"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"net"
)

func (m *Master) handleWorkerHandshake(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	logrus.Debugf("handleWorkerHandshake packet received from: %s", conn.RemoteAddr())
	p.Type = common.PACKET_TYPE__HANDSHAKE_ACK
	writer := bufio.NewWriter(conn)
	_ = m.writePacket(writer, p)
	err := writer.Flush()
	if err != nil {
		logrus.Errorf("handleWorkerHandshake err: %v", err)
	}
}

func (m *Master) handleWorkerRegister(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	logrus.Debugf("handleWorkerRegister packet received from: %s", conn.RemoteAddr())
	r := NewRouter()
	err := r.UnmarshalBinary(p.Payload)
	if err != nil {
		logrus.Errorf("handleWorkerRegister router.UnmarshalBinary err: %v", err)
	}
	for p := range r.Routes {
		if m.r.LockRoute(p) {
			go m.addRouteListener(ctx, p, conn)
			logrus.Infof("new route listener port: %d", p)
		} else {
			logrus.Infof("exist route listener port: %d", p)
		}
	}
	p.Type = common.PACKET_TYPE__REGISTER_ACK
	writer := bufio.NewWriter(conn)
	_ = m.writePacket(writer, p)
	err = writer.Flush()
	if err != nil {
		logrus.Errorf("handleWorkerRegister err: %v", err)
	}
}
