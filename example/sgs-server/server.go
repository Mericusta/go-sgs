package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/framework"
)

type Server struct {
	*framework.Framework
	userMgr *sync.Map
}

func NewServer() *Server {
	s := &Server{
		Framework: framework.New(),
		userMgr:   &sync.Map{},
	}
	s.AppendAcceptor(acceptor.NewServerAcceptor(
		"tcp", config.DefaultServerAddress, config.TcpKeepAlive,
	))
	s.AppendHandleMiddleware(
		NewServerMiddleware(s),
		NewUserMiddleware(s),
	)
	return s
}

func (s *Server) UserMgr() *sync.Map {
	return s.userMgr
}
