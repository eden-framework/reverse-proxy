package worker

import (
	"bufio"
	"github.com/eden-framework/context"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"net"
)

func (w *Worker) handleMasterConn(ctx *context.WaitStopContext, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	scanner.Split(common.PacketSplitFunc)

Run:
	for {
		select {
		case <-ctx.Done():
			break Run
		default:
			if scanner.Scan() {
				var packet = new(common.Packet)
				err := packet.UnmarshalBinary(scanner.Bytes())
				if err != nil {
					logrus.Warningf("invalid packet err: %v", err)
					continue
				}
				w.handleMasterPacket(ctx, packet, conn)
				if packet.Type == common.PACKET_TYPE__REGISTER_ACK {
					break Run
				}
			} else {
				break Run
			}
		}
	}
}

func (w *Worker) handleMasterPacket(ctx *context.WaitStopContext, p *common.Packet, conn net.Conn) {
	switch p.Type {
	case common.PACKET_TYPE__HANDSHAKE_ACK:
		w.handleMasterHandshakeAck(ctx, p, conn)
	case common.PACKET_TYPE__REGISTER_ACK:
		w.handleMasterRegisterAck(ctx, p, conn)
	}
}
