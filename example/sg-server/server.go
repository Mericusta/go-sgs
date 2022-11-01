package main

import (
	"sync"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/framework"
)

type SGServer struct {
	*framework.Framework
	userMgr *sync.Map
}

func NewSGServer() *SGServer {
	sgs := &SGServer{
		Framework: framework.New(),
		userMgr:   &sync.Map{},
	}
	sgs.AppendAcceptor(acceptor.NewServerAcceptor(
		"tcp", config.DefaultServerAddress, config.TcpKeepAlive,
	))
	sgs.AppendHandleMiddleware(
		NewServerMiddleware(sgs),
		NewUserMiddleware(sgs),
	)
	return sgs
}

func (s *SGServer) UserMgr() *sync.Map {
	return s.userMgr
}
