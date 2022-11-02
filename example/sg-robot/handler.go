package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

type RobotMgrHandler func(IRobotMgrContext, protocol.Protocol)

var robotMgrHandlerMrg map[protocol.ProtocolID]RobotMgrHandler

func RegisterRobotMgrHandler() {
	robotMgrHandlerMrg = make(map[protocol.ProtocolID]RobotMgrHandler)
	robotMgrHandlerMrg[msg.S2CMsgID_Login] = func(ctx IRobotMgrContext, p protocol.Protocol) {
		s2cMsg, ok := p.(*msg.S2CLoginData)
		if s2cMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Login, p)
			return
		}

		robot := model.NewRobot(ctx.Link().UID())
		robot.SetCounter(s2cMsg.User.GetCounter())
		ctx.RobotMgr().Store(ctx.Link().UID(), robot)

		fmt.Printf("Note: robot %v login with init counter %v\n", robot.ID(), s2cMsg.User.GetCounter())

		key := int(time.Now().UnixNano())
		v1, v2 := rand.Intn(1024), rand.Intn(1024)
		robot.AddExpect(key, v1+v2)
		c2sMsg := &msg.C2SBusinessData{
			Key: key, Value1: v1, Value2: v2,
		}
		ctx.Link().Send(event.New(protocol.ProtocolID(msg.C2SMsgID_Business), c2sMsg))
		fmt.Printf("Note: robot %v send business key %v value1 %v value2 %v, expect %v\n", robot.ID(), key, v1, v2, v1+v2)
	}
}

type RobotHandler func(IRobotContext, protocol.Protocol)

var robotHandlerMgr map[protocol.ProtocolID]RobotHandler

func RegisterRobotHandler() {
	robotHandlerMgr = make(map[protocol.ProtocolID]RobotHandler)
	robotHandlerMgr[msg.S2CMsgID_Business] = func(ctx IRobotContext, p protocol.Protocol) {
		s2cMsg, ok := p.(*msg.S2CBusinessData)
		if s2cMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.S2CMsgID_Business, p)
			return
		}

		if v, has := ctx.Robot().GetExpect(s2cMsg.Key); !has || v != s2cMsg.Result {
			fmt.Printf("Error: robot %v S2CMsgID_Business key %v result %v not match expect %v\n", ctx.Robot().ID(), s2cMsg.Key, s2cMsg.Result, v)
			return
		}

		ctx.Robot().DelExpect(s2cMsg.Key)
		fmt.Printf("Note: robot %v recv business key %v result %v, then delete expect\n", ctx.Robot().ID(), s2cMsg.Key, s2cMsg.Result)
	}
}
