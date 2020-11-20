package master

import (
	"bufio"
	"github.com/eden-framework/context"
	"github.com/robotic-framework/reverse-proxy/codec"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"net"
)

func (m *Master) handleWorkerConnection(ctx *context.WaitStopContext, conn net.Conn) {
	logrus.Infof("new worker connected: %s", conn.RemoteAddr())
	defer logrus.Infof("worker %s disconnected", conn.RemoteAddr())

	scanner := bufio.NewScanner(conn)
	scanner.Split(codec.InternalUnpack)

	for {
		if scanner.Scan() {
			var packet = new(common.Packet)
			err := packet.UnmarshalBinary(scanner.Bytes())
			if err != nil {
				logrus.Warningf("invalid packet err: %v", err)
				continue
			}
			m.handleWorkerPacket(ctx, packet, conn)
			if packet.Type == common.PACKET_TYPE__REGISTER_ACK {
				break
			}
		} else {
			break
		}
	}
}

func (m *Master) handleWorkerPacket(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	switch p.Type {
	case common.PACKET_TYPE__HANDSHAKE:
		m.handleWorkerHandshake(ctx, p, conn)
	case common.PACKET_TYPE__REGISTER:
		m.handleWorkerRegister(ctx, p, conn)
	}
}
