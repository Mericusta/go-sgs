package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/example/sgs-server/server"
	"github.com/Mericusta/go-sgs/example/sgs-server/user"
)

func main() {
	// register protocol ID handler
	server.RegisterHandler()   // use server context
	user.RegisterUserHandler() // use user context

	// create server
	sgs := server.NewServer()

	// append middleware
	sgs.AppendHandleMiddleware(
		server.NewMiddleware(sgs),
		user.NewMiddleware(sgs),
	)

	// run server
	go sgs.Run()

	// watch system signal
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	sgs.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
