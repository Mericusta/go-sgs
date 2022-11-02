package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/example/msg"
)

func main() {
	const robotCount int = 1

	// register msg ID protocol
	msg.Init()

	// register msg ID handler
	RegisterRobotMgrHandler()
	RegisterRobotHandler()

	// create robot
	fmt.Printf("Note: SG-Robot run\n")
	sgr := NewSGRobot(robotCount)

	// run robot
	go sgr.Run()

	// watch system signal
	s := make(chan os.Signal, 10)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: exit\n")
	sgr.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
