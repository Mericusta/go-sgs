package server

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/middleware"
)

type ServerMiddleware struct {
	server *Server
}

func NewMiddleware(server *Server) *ServerMiddleware {
	return &ServerMiddleware{server: server}
}

func (m *ServerMiddleware) Do(ctx middleware.IContext, e *event.Event) bool {
	if handler, has := serverHandlerMgr[e.ID()]; handler != nil && has {
		handler(NewContext(ctx, m.server), e.Data())
		return false
	}
	return true
}
