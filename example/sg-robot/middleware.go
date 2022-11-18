package main

import (
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/logger"
	"go.uber.org/zap"
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
		iRobot, has := m.sgRobot.RobotMgr().Load(ctx.Linker().UID()) // TODO: 性能瓶颈
		if !has {
			logger.Logger().Error("can not find robot by link", zap.Uint64("link", ctx.Linker().UID()))
			return false
		}
		robot, ok := iRobot.(*model.Robot)
		if !ok {
			logger.Logger().Error("robot manager link value type is not *Robot", zap.Uint64("link", ctx.Linker().UID()))
			return false
		}
		handler(NewRobotContext(ctx, robot), e.Data())
		return false
	}
	return true
}
