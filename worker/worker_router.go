package worker

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/eden-framework/context"
	"github.com/sirupsen/logrus"
	"net"
)

type Route struct {
	remotePort int
	handler    Handler

	conn net.Conn
}

func (r *Route) UnmarshalBinary(data []byte) error {
	r.remotePort = int(binary.BigEndian.Uint32(data))
	return nil
}

func (r Route) MarshalBinary() (data []byte, err error) {
	data = make([]byte, 4)
	binary.BigEndian.PutUint32(data, uint32(r.remotePort))
	return
}

func (r *Route) start(ctx *context.WaitStopContext, conn net.Conn) {
	r.conn = conn
	scanner := bufio.NewScanner(conn)
	scanner.Split(bufio.SplitFunc(r.handler.UnpackFunc))
	for scanner.Scan() {
		resp, err := r.handler.HandleFunc(scanner.Bytes())
		if err != nil {
			logrus.Errorf("route %d handle err: %v", r.remotePort, err)
			continue
		}

		if resp != nil && len(resp) > 0 {
			resp, err = r.handler.PackFunc(resp)
			if err != nil {
				logrus.Errorf("route %d pack resp err: %v", r.remotePort, err)
				continue
			}
			err = r.Send(resp)
			if err != nil {
				logrus.Errorf("route %d write resp err: %v", r.remotePort, err)
			}
		}
	}
}

func (r *Route) stop() {
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *Route) Send(payload []byte) error {
	_, err := r.conn.Write(payload)
	return err
}

type Router struct {
	Routes map[int]Route
}

func NewRouter() *Router {
	return &Router{
		Routes: make(map[int]Route),
	}
}

func (r *Router) Close() {
	for _, route := range r.Routes {
		route.stop()
	}
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

		route := Route{}
		err = route.UnmarshalBinary(routePortBytes)
		if err != nil {
			return err
		}

		r.Routes[route.remotePort] = route
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

func (r *Router) AddRoute(remotePort int, handler Handler) *Route {
	route := Route{
		remotePort: remotePort,
		handler:    handler,
	}
	r.Routes[remotePort] = route
	return &route
}

func (r *Router) GetRoute(remotePort int) *Route {
	if v, ok := r.Routes[remotePort]; ok {
		return &v
	}
	return nil
}
