package main

import (
	"github.com/Mericusta/go-sgs/example/model/robot"
	"github.com/Mericusta/go-sgs/middleware"
)

type IRobotContext interface {
	middleware.IContext
	Robot() *robot.Robot
}

type RobotContext struct {
	middleware.IContext
	r *robot.Robot
}

func NewRobotContext(ctx middleware.IContext, r *robot.Robot) IRobotContext {
	return &RobotContext{
		IContext: ctx,
		r:        r,
	}
}

func (ctx *RobotContext) Robot() *robot.Robot {
	return ctx.r
}
