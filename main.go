package main

import (
	"ahs_server/transfer"
	"os"
	"os/signal"
)

import (
	_ "ahs_server/transfer/quic"
	_ "ahs_server/transfer/tcp"
)

func main() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	transfer.Init()
	<-quit
	transfer.Stop()
}
