package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

type RobotHandler func(IRobotContext, protocol.Protocol)

var robotHandlerMgr map[protocol.ProtocolID]RobotHandler

func RegisterRobotHandler() {
	robotHandlerMgr = make(map[protocol.ProtocolID]RobotHandler)
	robotHandlerMgr[msg.S2CMsgID_Login] = func(ctx IRobotContext, p protocol.Protocol) {
		s2cMsg, ok := p.(*msg.S2CLoginData)
		if s2cMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Login, p)
			return
		}

		ctx.Robot().SetCounter(s2cMsg.User.GetCounter())
	}
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
	}
}
