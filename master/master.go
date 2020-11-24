package master

import (
	context2 "context"
	"fmt"
	"github.com/eden-framework/context"
	"github.com/eden-framework/reverse-proxy/codec"
	"github.com/eden-framework/reverse-proxy/common"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync/atomic"
)

type Master struct {
	ListenAddr string

	sequence uint32
	r        *Router
	listener net.Listener
	ctx      *context.WaitStopContext
}

func (m *Master) Init() {
	m.setDefaults()
	m.r = NewRouter()
}

func (m *Master) setDefaults() {
	m.sequence = 0
}

func (m *Master) Start(ctx *context.WaitStopContext) {
	if ctx == nil {
		ctx = context.NewWaitStopContext()
		m.ctx = ctx
	}

	ctx.Add(1)
	defer ctx.Finish()

	// 监听worker的连接请求
	var err error
	m.listener, err = net.Listen("tcp4", m.ListenAddr)
	if err != nil {
		panic(err)
	}

	logrus.Infof("master server start listening at %s", m.ListenAddr)
	defer logrus.Info("master server stopped")

	go func() {
		for {
			conn, err := m.listener.Accept()
			if err != nil {
				logrus.Warningf("master accept err: %v", err)
				break
			}
			go m.handleWorkerConnection(ctx, conn)
		}
	}()

	<-ctx.Done()
	_ = m.listener.Close()
	m.r.Close()
}

func (m *Master) Stop() {
	if m.ctx != nil {
		m.ctx.Cancel()
	}
}

func (m *Master) writePacket(writer io.Writer, p *common.Packet) error {
	if p.Sequence == 0 {
		atomic.AddUint32(&m.sequence, 1)
		p.Sequence = m.sequence
	}
	packetBytes, err := p.MarshalBinary()
	packetBytes, err = codec.InternalPack(packetBytes)
	if err != nil {
		return err
	}

	_, err = writer.Write(packetBytes)
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
