package main

import (
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/logger"
)

const robotCount int = 1
const requestCount int = 10

func main() {
	// register msg ID protocol
	msg.Init()

	// register msg ID handler
	RegisterRobotMgrHandler()
	RegisterRobotHandler()

	// create robot
	sgr := NewSGRobot(robotCount)

	// run robot
	logger.Logger().Info("SG-Robot run")
	go sgr.Run()

	// hold robot
	sgr.Hold()
}
