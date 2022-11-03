package main

import (
	"github.com/Mericusta/go-logger"
	"github.com/Mericusta/go-sgs/example/msg"
)

func main() {
	const robotCount int = 1
	logger.Init(robotCount, "std")

	// register msg ID protocol
	msg.Init()

	// register msg ID handler
	RegisterRobotMgrHandler()
	RegisterRobotHandler()

	// create robot
	sgr := NewSGRobot(robotCount)

	// run robot
	logger.Info().Package("main").Content("SG-Robot run")
	go sgr.Run()

	// hold robot
	sgr.Hold()
}
