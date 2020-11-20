package master

import (
	"bytes"
	context2 "context"
	"encoding/binary"
	"fmt"
	"github.com/eden-framework/context"
	"github.com/sirupsen/logrus"
	"net"
)

type Route struct {
	RemotePort int

	sourceConn net.Conn
	targetConn net.Conn

	sourceReaderStarted, targetReaderStarted bool
}

func (r *Route) UnmarshalBinary(data []byte) error {
	r.RemotePort = int(binary.BigEndian.Uint32(data))
	return nil
}

func (r Route) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(r.RemotePort))
	return
}

func (r *Route) startSourceReader() {
	if r.sourceReaderStarted {
		return
	}

	logrus.Debugf("route %d start source reading...", r.RemotePort)
	defer func() {
		logrus.Debugf("route %d finish source reading", r.RemotePort)

		_ = r.sourceConn.Close()
		r.sourceReaderStarted = false
	}()

	r.sourceReaderStarted = true

	for {
		buf := make([]byte, 4096)
		n, err := r.sourceConn.Read(buf)
		if err != nil {
			logrus.Errorf("proxy forward from source err: %v", err)
			break
		}
		_, err = r.targetConn.Write(buf[:n])
		if err != nil {
			logrus.Errorf("proxy forward to target err: %v", err)
			break
		}
	}
}

func (r *Route) startTargetReader(cancelFunc context2.CancelFunc) {
	if r.targetReaderStarted {
		return
	}

	logrus.Debugf("route %d start target reading...", r.RemotePort)
	defer func() {
		logrus.Debugf("route %d finish target reading", r.RemotePort)

		_ = r.targetConn.Close()
		r.targetReaderStarted = false
	}()

	r.targetReaderStarted = true

	for {
		buf := make([]byte, 4096)
		n, err := r.targetConn.Read(buf)
		if err != nil {
			logrus.Errorf("proxy forward from target err: %v", err)
			break
		}
		_, err = r.sourceConn.Write(buf[:n])
		if err != nil {
			logrus.Errorf("proxy forward to source err: %v", err)
			break
		}
	}
	if cancelFunc != nil {
		cancelFunc()
	}
}

/**

 */
func (r *Route) Start(ctx *context.WaitStopContext) {
	go r.startSourceReader()
	go r.startTargetReader(ctx.TempCancel())
}

func (r *Route) Close() {
	if r.targetConn != nil {
		_ = r.targetConn.Close()
	}
	if r.sourceConn != nil {
		_ = r.sourceConn.Close()
	}
}

type Router struct {
	Routes map[int]*Route
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[int]*Route),
	}
}

func (r *Router) Close() {
	for _, route := range r.Routes {
		route.Close()
	}
}

func (r *Router) LockRoute(remotePort int) bool {
	if _, ok := r.Routes[remotePort]; ok {
		return false
	}
	r.Routes[remotePort] = &Route{
		RemotePort: remotePort,
	}
	return true
}

func (r *Router) ReleaseRoute(remotePort int) {
	if _, ok := r.Routes[remotePort]; ok {
		delete(r.Routes, remotePort)
	}
}

func (r *Router) BindRoute(ctx *context.WaitStopContext, remotePort int, sourceConn, targetConn net.Conn) {
	var route *Route
	var ok bool
	if route, ok = r.Routes[remotePort]; ok {
		route.sourceConn = sourceConn
		route.targetConn = targetConn
	} else {
		route = &Route{
			RemotePort: remotePort,
			sourceConn: sourceConn,
			targetConn: targetConn,
		}
		r.Routes[remotePort] = route
	}
	go route.Start(ctx)
}

func (r *Router) UnmarshalBinary(data []byte) error {
	if len(r.Routes) > 0 {
		return fmt.Errorf("router is not empty")
	}
	reader := bytes.NewReader(data)

	routeLengthBytes := make([]byte, 4)
	_, err := reader.Read(routeLengthBytes)
	if err != nil {
		return err
	}

	routeLength := int(binary.BigEndian.Uint32(routeLengthBytes))
	for i := 0; i < routeLength; i++ {
		routePortBytes := make([]byte, 4)
		_, err = reader.Read(routePortBytes)
		if err != nil {
			return err
		}

		route := &Route{}
		err = route.UnmarshalBinary(routePortBytes)
		if err != nil {
			return err
		}

		r.Routes[route.RemotePort] = route
	}
	return nil
}

func (r Router) MarshalBinary() (data []byte, err error) {
	buf := bytes.NewBuffer([]byte{})

	routeLengthBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(routeLengthBytes, uint32(len(r.Routes)))
	_, err = buf.Write(routeLengthBytes)
	if err != nil {
		return
	}

	for _, route := range r.Routes {
		routeBytes, err := route.MarshalBinary()
		if err != nil {
			return nil, err
		}
		_, err = buf.Write(routeBytes)
		if err != nil {
			return nil, err
		}
	}
	data = buf.Bytes()
	return
}
