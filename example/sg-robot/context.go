package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/example/model/robot"
)

type IRobotMgrContext interface {
	dispatcher.IContext
	RobotMgr() *sync.Map
}

type RobotMgrContext struct {
	dispatcher.IContext
	*SGRobot
}

func NewRobotMgrContext(ctx dispatcher.IContext, sgRobot *SGRobot) IRobotMgrContext {
	return &RobotMgrContext{
		IContext: ctx,
		SGRobot:  sgRobot,
	}
}

type IRobotContext interface {
	dispatcher.IContext
	Robot() *robot.Robot
}

type RobotContext struct {
	dispatcher.IContext
	r *robot.Robot
}

func NewRobotContext(ctx dispatcher.IContext, r *robot.Robot) IRobotContext {
	return &RobotContext{
		IContext: ctx,
		r:        r,
	}
}

func (ctx *RobotContext) Robot() *robot.Robot {
	return ctx.r
}
