package server

import (
	"sync"

	"github.com/Mericusta/go-sgs/framework"
)

type Server struct {
	*framework.Framework
	userMgr *sync.Map
}

func NewServer() *Server {
	return &Server{
		Framework: framework.New(),
		userMgr:   &sync.Map{},
	}
}

func (s *Server) UserMgr() *sync.Map {
	return s.userMgr
}
