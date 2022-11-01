package main

import (
	"fmt"
	"sync"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/protocol"
)

type robotRunMiddleware struct {
	wg *sync.WaitGroup
}

func NewRobotRunMiddleware(wg *sync.WaitGroup) *robotRunMiddleware {
	return &robotRunMiddleware{wg: wg}
}

func (rmd *robotRunMiddleware) Do(ctx dispatcher.IContext) bool {
	fmt.Printf("Note: robot %v dial done\n", ctx.Link().UID())
	rmd.wg.Done()
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
			"tcp", config.DefaultServerAddress, config.TcpDialOvertime,
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

func (rm *SGRobot) RobotMgr() *sync.Map {
	return rm.robotMgr
}

func (rm *SGRobot) Run() {
	go rm.Framework.Run()
	rm.dialWaitGroup.Wait()
	rm.ForRangeDispatcher(func(u uint64, d *dispatcher.Dispatcher) bool {
		fmt.Printf("Note: robot %v send event %v\n", u, msg.C2SMsgID_Login)
		d.Send(event.New(
			protocol.ProtocolID(msg.C2SMsgID_Login),
			&msg.C2SLoginData{AccountID: u},
		))
		return true
	})
}

func (rm *SGRobot) Exit() {

}
