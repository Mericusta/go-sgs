package user

import (
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/middleware"
)

type IUserContext interface {
	middleware.IContext
	User() *serverModel.User
}

type UserContext struct {
	middleware.IContext
	user *serverModel.User
}

func NewContext(ctx middleware.IContext, user *serverModel.User) IUserContext {
	return &UserContext{
		IContext: ctx,
		user:     user,
	}
}

func (ctx *UserContext) User() *serverModel.User {
	return ctx.user
}
