package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var (
	clientMode = flag.Bool("client", false, "client mode")
	bindAddr   = flag.String("bind-addr", ":6000", "master bind address")
	masterAddr = flag.String("master-addr", "", "master address")
	remotePort = flag.Int("remote-port", 0, "remote proxy port")
	localAddr  = flag.String("local-addr", "", "local addr")
)

func main() {
	flag.Parse()

	runtime.GOMAXPROCS(1)

	if *clientMode {
		slave := NewProxySlave()
		slave.Start(*masterAddr, uint16(*remotePort), *localAddr)
	} else {
		master := NewProxyMaster()
		master.Start(*bindAddr)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	<-signalChan
}
