package user

import (
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

type UserHandler func(IUserContext, protocol.Protocol)

var userHandlerMgr map[protocol.ProtocolID]UserHandler

func RegisterUserHandler() {
	userHandlerMgr = make(map[protocol.ProtocolID]UserHandler)
	userHandlerMgr[msg.C2SMsgID_Business] = func(ctx IUserContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SBusinessData)
		if c2sMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Business, p)
			return
		}

		ctx.User().AddCounter()

		s2cMsg := &msg.S2CBusinessData{
			Key:    c2sMsg.Key,
			Result: c2sMsg.Value1 + c2sMsg.Value2,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Business, s2cMsg))
	}
}
