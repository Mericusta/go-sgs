package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"
)

func main() {
	// register protocol ID handler
	RegisterHandler()     // use server context
	RegisterUserHandler() // use user context

	// create server
	sgs := NewServer()

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
