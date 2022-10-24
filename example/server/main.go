package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/framework"
)

type UserMgr struct {
	*framework.Framework
	userMgr sync.Map
}

func NewUserMgr(ctx context.Context) *UserMgr {
	return &UserMgr{
		Framework: framework.New(),
		userMgr:   sync.Map{},
	}
}

func main() {
	// register server protocol ID handler
	registerServerMsgCallback()

	// create server
	serverCtx, serverCanceler := context.WithCancel(context.Background())
	server := NewUserMgr(serverCtx)

	// run server
	go server.Run(serverCtx)

	// watch system signal
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
