package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/example/model"
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
	User() *model.User
}

type UserContext struct {
	dispatcher.IContext
	user *model.User
}

func NewUserContext(ctx dispatcher.IContext, user *model.User) IUserContext {
	return &UserContext{
		IContext: ctx,
		user:     user,
	}
}

func (ctx *UserContext) User() *model.User {
	return ctx.user
}
