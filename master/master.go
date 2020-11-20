package master

import (
	"bytes"
	context2 "context"
	"encoding/binary"
	"fmt"
	"github.com/eden-framework/context"
	"github.com/robotic-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync/atomic"
)

type Master struct {
	ListenAddr string

	sequence uint32
	r        *Router
}

func (m *Master) Init() {
	m.setDefaults()
	m.r = NewRouter()
}

func (m *Master) setDefaults() {
	m.sequence = 0
}

func (m *Master) Start(ctx *context.WaitStopContext) {
	// 监听worker的连接请求
	listener, err := net.Listen("tcp4", m.ListenAddr)
	if err != nil {
		panic(err)
	}

	logrus.Infof("master server start listening at %s", m.ListenAddr)
	defer logrus.Info("master server stopped")

	go func() {
		ctx.Add(1)

		<-ctx.Done()
		_ = listener.Close()
		m.r.Close()

		ctx.Finish()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Warningf("master accept err: %v", err)
			break
		}
		go m.handleWorkerConnection(ctx, conn)
	}
}

func (m *Master) writePacket(writer io.Writer, p *common.Packet) error {
	if p.Sequence == 0 {
		atomic.AddUint32(&m.sequence, 1)
		p.Sequence = m.sequence
	}
	packetBytes, err := p.MarshalBinary()
	if err != nil {
		return err
	}
	packetLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(packetLengthBytes, uint32(len(packetBytes)))

	buf := bytes.NewBuffer([]byte{})
	buf.WriteString(common.PacketBytesPrefix)
	buf.Write(packetLengthBytes)
	buf.Write(packetBytes)

	_, err = writer.Write(buf.Bytes())
	return err
}

func (m *Master) addRouteListener(globalCtx *context.WaitStopContext, remotePort int, targetConn net.Conn) {
	listener, err := net.Listen("tcp4", fmt.Sprintf(":%d", remotePort))
	if err != nil {
		logrus.Errorf("addRouteListener listen err: %v", err)
		return
	}

	logrus.Debugf("route %d start listening...", remotePort)
	defer func() {
		logrus.Debugf("route %d finish listening", remotePort)
		m.r.ReleaseRoute(remotePort)
	}()

	routeCtx, cancel := context2.WithCancel(context2.Background())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				logrus.Warningf("addRouteListener accept err: %v", err)
				break
			}
			context.WithTempCancel(globalCtx, cancel)
			m.r.BindRoute(globalCtx, remotePort, conn, targetConn)
		}
	}()

	<-routeCtx.Done()
	_ = listener.Close()
}
