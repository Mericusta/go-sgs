package server

import (
	"sync"

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

func NewContext(ctx middleware.IContext, server *Server) IServerContext {
	return &ServerContext{
		IContext: ctx,
		Server:   server,
	}
}
