package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type robotRunMiddleware struct {
	wg *sync.WaitGroup
}

func NewRobotRunMiddleware(wg *sync.WaitGroup) *robotRunMiddleware {
	return &robotRunMiddleware{wg: wg}
}

func (rrm *robotRunMiddleware) Do(ctx dispatcher.IContext) bool {
	logger.Log().Info("robot dial done", zap.Uint64("linker", ctx.UID()))
	rrm.wg.Done()
	return true
}

type SGRobot struct {
	*framework.Framework
	robotMgr      *sync.Map
	dialWaitGroup *sync.WaitGroup
}

func NewSGRobot(count int) *SGRobot {
	sgr := &SGRobot{
		Framework:     framework.New(),
		robotMgr:      &sync.Map{},
		dialWaitGroup: &sync.WaitGroup{},
	}
	for index := 0; index != count; index++ {
		sgr.AppendAcceptor(acceptor.NewClientAcceptor(
			index, "tcp", config.DefaultServerAddress, config.TcpDialOvertime,
		))
	}
	sgr.AppendHandleMiddleware(
		NewRobotMgrMiddleware(sgr),
		NewRobotMiddleware(sgr),
	)
	sgr.dialWaitGroup.Add(count)
	sgr.SetRunMiddleware(NewRobotRunMiddleware(sgr.dialWaitGroup))
	return sgr
}

func (sgr *SGRobot) RobotMgr() *sync.Map {
	return sgr.robotMgr
}

func (sgr *SGRobot) Run() {
	go sgr.Framework.Run()
	sgr.dialWaitGroup.Wait()
	sgr.ForRangeDispatcher(func(u uint64, d *dispatcher.Dispatcher) bool {
		logger.Log().Info("robot send event", zap.Int("dispatcher", d.Index()), zap.Uint64("linker", u), zap.Int("ID", msg.C2SMsgID_Login))
		d.Send(event.New(
			u, protocol.ProtocolID(msg.C2SMsgID_Login), &msg.C2SLoginData{AccountID: u},
		))
		return true
	})
}
