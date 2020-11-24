package main

import (
	"github.com/eden-framework/context"
	master2 "github.com/eden-framework/reverse-proxy/master"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.NewWaitStopContext()
	addr := "127.0.0.1:9067"

	master := &master2.Master{
		ListenAddr: addr,
	}
	master.Init()
	go master.Start(ctx)

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
