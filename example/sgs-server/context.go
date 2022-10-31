package main

import (
	"sync"

	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/middleware"
)

type IServerContext interface {
	middleware.IContext
	UserMgr() *sync.Map
}

type ServerContext struct {
	middleware.IContext
	*Server
}

func NewServerContext(ctx middleware.IContext, server *Server) IServerContext {
	return &ServerContext{
		IContext: ctx,
		Server:   server,
	}
}

type IUserContext interface {
	middleware.IContext
	User() *serverModel.User
}

type UserContext struct {
	middleware.IContext
	user *serverModel.User
}

func NewUserContext(ctx middleware.IContext, user *serverModel.User) IUserContext {
	return &UserContext{
		IContext: ctx,
		user:     user,
	}
}

func (ctx *UserContext) User() *serverModel.User {
	return ctx.user
}
