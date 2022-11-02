package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
)

type RobotMgrMiddleware struct {
	sgRobot *SGRobot
}

func NewRobotMgrMiddleware(sgRobot *SGRobot) *RobotMgrMiddleware {
	return &RobotMgrMiddleware{sgRobot: sgRobot}
}

func (m *RobotMgrMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := robotMgrHandlerMrg[e.ID()]; handler != nil && has {
		handler(NewRobotMgrContext(ctx, m.sgRobot), e.Data())
		return false
	}
	return true
}

type RobotMiddleware struct {
	sgRobot *SGRobot
}

func NewRobotMiddleware(sgRobot *SGRobot) *RobotMiddleware {
	return &RobotMiddleware{sgRobot: sgRobot}
}

func (m *RobotMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := robotHandlerMgr[e.ID()]; handler != nil && has {
		iRobot, has := m.sgRobot.RobotMgr().Load(ctx.Link().UID()) // TODO: 性能瓶颈
		if !has {
			fmt.Printf("Error: can not find robot by uid %v", ctx.Link().UID())
			return false
		}
		robot, ok := iRobot.(*model.Robot)
		if !ok {
			fmt.Printf("Error: robot manager uid %v value type is not *Robot\n", ctx.Link().UID())
			return false
		}
		handler(NewRobotContext(ctx, robot), e.Data())
		return false
	}
	return true
}
