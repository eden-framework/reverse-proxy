package worker

import (
	"bufio"
	"github.com/eden-framework/context"
	"github.com/eden-framework/reverse-proxy/codec"
	"github.com/eden-framework/reverse-proxy/common"
	"github.com/profzone/envconfig"
	"github.com/sirupsen/logrus"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type Worker struct {
	RemoteAddr    string
	RetryInterval envconfig.Duration
	RetryMaxTime  int

	sequence uint32
	r        *Router
	ctx      *context.WaitStopContext
}

func (w *Worker) Init() {
	w.setDefaults()
	w.r = NewRouter()
}

func (w *Worker) setDefaults() {
	if w.RetryMaxTime == 0 {
		w.RetryMaxTime = 5
	}
	if w.RetryInterval == 0 {
		w.RetryInterval = envconfig.Duration(time.Second)
	}
}

func (w *Worker) Start(ctx *context.WaitStopContext) {
	if ctx == nil {
		ctx = context.NewWaitStopContext()
		w.ctx = ctx
	}
	ctx.Add(1)
	defer ctx.Finish()

	conn, err := net.Dial("tcp4", w.RemoteAddr)
	if err != nil {
		panic(err)
	}

	logrus.Infof("worker connected with master: %s", w.RemoteAddr)
	defer logrus.Info("worker stopped")

	go w.handleMasterConn(ctx, conn)

	writer := bufio.NewWriter(conn)
	err = w.handshake(writer)

	<-ctx.Done()
	w.r.Close()
	_ = conn.Close()
}

func (w *Worker) Stop() {
	if w.ctx != nil {
		w.ctx.Cancel()
	}
}

func (w *Worker) writePacket(writer io.Writer, p *common.Packet) error {
	if p.Sequence == 0 {
		atomic.AddUint32(&w.sequence, 1)
		p.Sequence = w.sequence
	}
	packetBytes, err := p.MarshalBinary()
	packetBytes, err = codec.InternalPack(packetBytes)
	if err != nil {
		return err
	}

	_, err = writer.Write(packetBytes)
	return err
}

func (w *Worker) AddRoute(remotePort int, handler Handler) *Route {
	return w.r.AddRoute(remotePort, handler)
}

func (w *Worker) GetRoute(remotePort int) *Route {
	return w.r.GetRoute(remotePort)
}
