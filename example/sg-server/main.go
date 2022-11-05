package main

import (
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/logger"
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
	logger.Logger().Info("SG-Server run")
	go sgs.Run()

	// hold server
	sgs.Hold()
}
