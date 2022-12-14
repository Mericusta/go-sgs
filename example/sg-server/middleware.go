package main

import (
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/logger"
	"go.uber.org/zap"
)

type ServerHandlerMiddleware struct {
	sgServer *SGServer
}

func NewServerHandlerMiddleware(sgServer *SGServer) *ServerHandlerMiddleware {
	return &ServerHandlerMiddleware{sgServer: sgServer}
}

func (m *ServerHandlerMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := serverHandlerMgr[e.ID()]; handler != nil && has {
		handler(NewServerContext(ctx, m.sgServer), e.Data())
		return false
	}
	return true
}

type UserHandlerMiddleware struct {
	sgServer *SGServer
}

func NewUserHandlerMiddleware(sgServer *SGServer) *UserHandlerMiddleware {
	return &UserHandlerMiddleware{sgServer: sgServer}
}

func (m *UserHandlerMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := userHandlerMgr[e.ID()]; handler != nil && has {
		iUser, has := m.sgServer.UserMgr().Load(ctx.UID()) // TODO: 性能瓶颈
		if !has {
			logger.Log().Error("can not find user by link", zap.Uint64("linker", ctx.UID()))
			return false
		}
		user, ok := iUser.(*model.User)
		if !ok {
			logger.Log().Error("server user manager link value type is not *User", zap.Uint64("linker", ctx.UID()))
			return false
		}
		handler(NewUserContext(ctx, user), e.Data())
		return false
	}
	return true
}
