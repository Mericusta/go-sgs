package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/middleware"
)

type ServerMiddleware struct {
	server *Server
}

func NewServerMiddleware(server *Server) *ServerMiddleware {
	return &ServerMiddleware{server: server}
}

func (m *ServerMiddleware) Do(ctx middleware.IContext, e *event.Event) bool {
	if handler, has := serverHandlerMgr[e.ID()]; handler != nil && has {
		handler(NewServerContext(ctx, m.server), e.Data())
		return false
	}
	return true
}

type UserMiddleware struct {
	server *Server
}

func NewUserMiddleware(server *Server) *UserMiddleware {
	return &UserMiddleware{server: server}
}

func (m *UserMiddleware) Do(ctx middleware.IContext, e *event.Event) bool {
	if handler, has := userHandlerMgr[e.ID()]; handler != nil && has {
		iUser, has := m.server.UserMgr().Load(ctx.Link().UID())
		if !has {
			fmt.Printf("Error: can not find user by uid %v", ctx.Link().UID())
			return false
		}
		user, ok := iUser.(*serverModel.User)
		if !ok {
			fmt.Printf("Error: server user manager uid %v value type is not *User\n", ctx.Link().UID())
			return false
		}
		handler(NewUserContext(ctx, user), e.Data())
		return false
	}
	return true
}
