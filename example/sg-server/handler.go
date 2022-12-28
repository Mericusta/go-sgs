package main

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type ServerHandler func(IServerContext, protocol.ProtocolMsg)

var serverHandlerMgr map[protocol.ProtocolID]ServerHandler

func RegisterHandler() {
	serverHandlerMgr = make(map[protocol.ProtocolID]ServerHandler)
	serverHandlerMgr[msg.C2SMsgID_Login] = Login
	serverHandlerMgr[msg.C2SMsgID_Logout] = Logout
}

func Login(ctx IServerContext, p protocol.ProtocolMsg) {
	c2sMsg, ok := p.(*msg.C2SLoginData)
	if c2sMsg == nil || !ok {
		logger.Log().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Login), zap.Any("data", p))
		return
	}

	iUser, exists := ctx.UserMgr().LoadOrStore(ctx.UID(), model.NewUser())
	if exists {
		logger.Log().Warn("server user manager link already exists", zap.Uint64("linker", ctx.UID()))
	}
	user, ok := iUser.(*model.User)
	if !ok {
		logger.Log().Error("server user manager uid value type is not *model.User", zap.Uint64("linker", ctx.UID()))
		return
	}

	logger.Log().Info("link as user login with counter", zap.Uint64("linker", ctx.UID()), zap.Int("counter", user.Counter()))

	s2cMsg := &msg.S2CLoginData{
		User: &msg.User{
			Counter: user.Counter(),
		},
	}
	ctx.Send(event.New(ctx.UID(), msg.S2CMsgID_Login, s2cMsg))
}

func Logout(ctx IServerContext, p protocol.ProtocolMsg) {
	c2sMsg, ok := p.(*msg.C2SLogout)
	if c2sMsg == nil || !ok {
		logger.Log().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Logout), zap.Any("data", p))
		return
	}

	logger.Log().Info("link as user logout", zap.Uint64("linker", ctx.UID()))

	ctx.Exit()
	ctx.UserMgr().Delete(ctx.UID())
}

type UserHandler func(IUserContext, protocol.ProtocolMsg)

var userHandlerMgr map[protocol.ProtocolID]UserHandler

func RegisterUserHandler() {
	userHandlerMgr = make(map[protocol.ProtocolID]UserHandler)
	userHandlerMgr[msg.C2SMsgID_Business] = Business
}

func Business(ctx IUserContext, p protocol.ProtocolMsg) {
	c2sMsg, ok := p.(*msg.C2SBusinessData)
	if c2sMsg == nil || !ok {
		logger.Log().Error("msg ID data not match", zap.Int("ID", msg.C2SMsgID_Business), zap.Any("data", p))
		return
	}

	ctx.User().CounterIncrease()
	logger.Log().Info("link as user recv business key value1 value2", zap.Uint64("linker", ctx.UID()), zap.Int("key", c2sMsg.Key), zap.Int("value1", c2sMsg.Value1), zap.Int("value2", c2sMsg.Value2))

	s2cMsg := &msg.S2CBusinessData{
		Key: c2sMsg.Key, Result: c2sMsg.Value1 + c2sMsg.Value2,
	}
	ctx.Send(event.New(ctx.UID(), msg.S2CMsgID_Business, s2cMsg))

	if ctx.User().Counter() == controlCount {
		panic("server panic here")
	}

	// time.Sleep(time.Second * 10)
	// // condition: server exit actively
	// ctx.Linker().Exit()
}
