package main

import (
	"github.com/eden-framework/context"
	worker2 "github.com/eden-framework/reverse-proxy/worker"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.NewWaitStopContext()
	addr := "127.0.0.1:9067"

	worker := &worker2.Worker{
		RemoteAddr: addr,
	}
	worker.Init()

	worker.AddRoute(19000, worker2.Handler{
		HandleFunc: handlePort19000,
		UnpackFunc: nil,
	})

	go worker.Start(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL)

	s := <-sig
	logrus.Infof("signal %s received", s.String())
	ctx.Cancel()
}

func handlePort19000(payload []byte) (response []byte, err error) {
	return
}
