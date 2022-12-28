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
	sgsAcceptor := acceptor.NewServerAcceptor(
		"tcp", config.DefaultServerAddress, config.TcpKeepAlive,
	)
	sgs.AppendAcceptor(sgsAcceptor)
	sgs.AppendHandleMiddleware(
		NewServerHandlerMiddleware(sgs),
		NewUserHandlerMiddleware(sgs),
	)
	return sgs
}

func (s *SGServer) UserMgr() *sync.Map {
	return s.userMgr
}
