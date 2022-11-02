package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/example/model"
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
	Robot() *model.Robot
}

type RobotContext struct {
	dispatcher.IContext
	r *model.Robot
}

func NewRobotContext(ctx dispatcher.IContext, r *model.Robot) IRobotContext {
	return &RobotContext{
		IContext: ctx,
		r:        r,
	}
}

func (ctx *RobotContext) Robot() *model.Robot {
	return ctx.r
}
