package main

import (
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/model"
	"github.com/Mericusta/go-sgs/logger"
	"go.uber.org/zap"
)

type ServerMiddleware struct {
	sgServer *SGServer
}

func NewServerMiddleware(sgServer *SGServer) *ServerMiddleware {
	return &ServerMiddleware{sgServer: sgServer}
}

func (m *ServerMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := serverHandlerMgr[e.ID()]; handler != nil && has {
		handler(NewServerContext(ctx, m.sgServer), e.Data())
		return false
	}
	return true
}

type UserMiddleware struct {
	sgServer *SGServer
}

func NewUserMiddleware(sgServer *SGServer) *UserMiddleware {
	return &UserMiddleware{sgServer: sgServer}
}

func (m *UserMiddleware) Do(ctx dispatcher.IContext, e *event.Event) bool {
	if handler, has := userHandlerMgr[e.ID()]; handler != nil && has {
		iUser, has := m.sgServer.UserMgr().Load(ctx.Linker().UID()) // TODO: 性能瓶颈
		if !has {
			logger.Logger().Error("can not find user by link", zap.Uint64("link", ctx.Linker().UID()))
			return false
		}
		user, ok := iUser.(*model.User)
		if !ok {
			logger.Logger().Error("server user manager link value type is not *User", zap.Uint64("link", ctx.Linker().UID()))
			return false
		}
		handler(NewUserContext(ctx, user), e.Data())
		return false
	}
	return true
}
