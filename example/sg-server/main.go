package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/example/msg"
)

func main() {
	// register msg ID protocol
	msg.Init()

	// register msg ID handler
	RegisterHandler()     // use server context
	RegisterUserHandler() // use user context

	// create server
	sgs := NewSGServer()

	// run server
	fmt.Printf("Note: SG-Server run\n")
	go sgs.Run()

	// watch system signal
	s := make(chan os.Signal, 10)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: exit\n")
	sgs.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
