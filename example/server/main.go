package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/framework"
)

func main() {
	// register server protocol ID handler
	registerServerMsgCallback()

	// create server
	serverCtx, serverCanceler := context.WithCancel(context.Background())
	server := framework.New()

	// run server
	go server.Run(serverCtx)

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	server.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: execute canceler\n")
	serverCanceler() // end logic goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
