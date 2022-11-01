package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/dispatcher"
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
)

type IServerContext interface {
	dispatcher.IContext
	UserMgr() *sync.Map
}

type ServerContext struct {
	dispatcher.IContext
	*SGServer
}

func NewServerContext(ctx dispatcher.IContext, sgServer *SGServer) IServerContext {
	return &ServerContext{
		IContext: ctx,
		SGServer: sgServer,
	}
}

type IUserContext interface {
	dispatcher.IContext
	User() *serverModel.User
}

type UserContext struct {
	dispatcher.IContext
	user *serverModel.User
}

func NewUserContext(ctx dispatcher.IContext, user *serverModel.User) IUserContext {
	return &UserContext{
		IContext: ctx,
		user:     user,
	}
}

func (ctx *UserContext) User() *serverModel.User {
	return ctx.user
}
