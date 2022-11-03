package main

import (
	"github.com/Mericusta/go-logger"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
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
			logger.Error().Package("main").Content("msg ID %v data %+v not match", msg.C2SMsgID_Login, p)
			return
		}

		iUser, exists := ctx.UserMgr().LoadOrStore(ctx.Link().UID(), model.NewUser())
		if exists {
			logger.Warn().Package("main").Content("server user manager uid %v already exists", ctx.Link().UID())
		}
		user, ok := iUser.(*model.User)
		if !ok {
			logger.Error().Package("main").Content("server user manager uid %v value type is not *model.User", ctx.Link().UID())
			return
		}

		logger.Info().Package("main").Content("user %v login with counter %v", ctx.Link().UID(), user.GetCounter())

		s2cMsg := &msg.S2CLoginData{
			User: user,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Login, s2cMsg))
	}
	serverHandlerMgr[msg.C2SMsgID_Logout] = func(ctx IServerContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SLogout)
		if c2sMsg == nil || !ok {
			logger.Error().Package("main").Content("msg ID %v data %+v not match", msg.C2SMsgID_Logout, p)
			return
		}

		ctx.UserMgr().Delete(ctx.Link().UID())
		logger.Info().Package("main").Content("user %v logout", ctx.Link().UID())
	}
}

type UserHandler func(IUserContext, protocol.Protocol)

var userHandlerMgr map[protocol.ProtocolID]UserHandler

func RegisterUserHandler() {
	userHandlerMgr = make(map[protocol.ProtocolID]UserHandler)
	userHandlerMgr[msg.C2SMsgID_Business] = func(ctx IUserContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SBusinessData)
		if c2sMsg == nil || !ok {
			logger.Error().Package("main").Content("msg ID %v data %+v not match", msg.C2SMsgID_Business, p)
			return
		}

		ctx.User().AddCounter()
		logger.Info().Package("main").Content("user %v recv business key %v value1 %v value2 %v", ctx.Link().UID(), c2sMsg.Key, c2sMsg.Value1, c2sMsg.Value2)

		s2cMsg := &msg.S2CBusinessData{
			Key: c2sMsg.Key, Result: c2sMsg.Value1 + c2sMsg.Value2,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Business, s2cMsg))
	}
}
