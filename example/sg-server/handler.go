package main

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type ServerHandler func(IServerContext, protocol.Protocol)

var serverHandlerMgr map[protocol.ProtocolID]ServerHandler

func RegisterHandler() {
	serverHandlerMgr = make(map[protocol.ProtocolID]ServerHandler)
	serverHandlerMgr[msg.C2SMsgID_Login] = func(ctx IServerContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SLoginData)
		if c2sMsg == nil || !ok {
			logger.Logger().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Login), zap.Any("data", p))
			return
		}

		iUser, exists := ctx.UserMgr().LoadOrStore(ctx.Link().UID(), model.NewUser())
		if exists {
			logger.Logger().Warn("server user manager link already exists", zap.Uint64("link", ctx.Link().UID()))
		}
		user, ok := iUser.(*model.User)
		if !ok {
			logger.Logger().Error("server user manager uid value type is not *model.User", zap.Uint64("link", ctx.Link().UID()))
			return
		}

		logger.Logger().Info("link as user login with counter", zap.Uint64("link", ctx.Link().UID()), zap.Int("counter", user.GetCounter()))

		s2cMsg := &msg.S2CLoginData{
			User: user,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Login, s2cMsg))
	}
	serverHandlerMgr[msg.C2SMsgID_Logout] = func(ctx IServerContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SLogout)
		if c2sMsg == nil || !ok {
			logger.Logger().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Logout), zap.Any("data", p))
			return
		}

		logger.Logger().Info("link as user logout", zap.Uint64("link", ctx.Link().UID()))

		ctx.Link().Exit()
		ctx.UserMgr().Delete(ctx.Link().UID())
	}
}

type UserHandler func(IUserContext, protocol.Protocol)

var userHandlerMgr map[protocol.ProtocolID]UserHandler

func RegisterUserHandler() {
	userHandlerMgr = make(map[protocol.ProtocolID]UserHandler)
	userHandlerMgr[msg.C2SMsgID_Business] = func(ctx IUserContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SBusinessData)
		if c2sMsg == nil || !ok {
			logger.Logger().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Business), zap.Any("data", p))
			return
		}

		ctx.User().AddCounter()
		logger.Logger().Info("link as user recv business key value1 value2", zap.Uint64("link", ctx.Link().UID()), zap.Int("key", c2sMsg.Key), zap.Int("value1", c2sMsg.Value1), zap.Int("value2", c2sMsg.Value2))

		s2cMsg := &msg.S2CBusinessData{
			Key: c2sMsg.Key, Result: c2sMsg.Value1 + c2sMsg.Value2,
		}
		ctx.Link().Send(event.New(msg.S2CMsgID_Business, s2cMsg))
	}
}
