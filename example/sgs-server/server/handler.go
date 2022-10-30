package server

import (
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

type ServerHandler func(IServerContext, protocol.Protocol)

var serverHandlerMgr map[protocol.ProtocolID]ServerHandler

func RegisterHandler() {
	serverHandlerMgr = make(map[protocol.ProtocolID]ServerHandler)
	serverHandlerMgr[msg.C2SMsgID_Login] = func(ctx IServerContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SLoginData)
		if c2sMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Login, p)
			return
		}

		iUser, exists := ctx.UserMgr().LoadOrStore(ctx.Link().UID(), serverModel.NewUser())
		if exists {
			fmt.Printf("Warn: server user manager uid %v already exists\n", ctx.Link().UID())
		}
		user, ok := iUser.(*serverModel.User)
		if !ok {
			fmt.Printf("Error: server user manager uid %v value type is not *serverModel.User\n", ctx.Link().UID())
			return
		}

		s2cMsg := &msg.S2CLoginData{
			AccountID: int(ctx.Link().UID()),
			User:      user,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Login, s2cMsg))
	}
}
