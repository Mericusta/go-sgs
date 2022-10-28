package main

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/Mericusta/go-sgs/event"
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

type Server struct {
	*framework.Framework
	userMgr sync.Map
}

func NewServer() *Server {
	return &Server{
		Framework: framework.New(),
		userMgr:   sync.Map{},
	}
}

type ServerContext interface {
	Link() *link.Link
	Send(*event.Event)
	AddUser(uint64, *serverModel.User)
	DelUser(uint64)
	GetUser(uint64) *serverModel.User
}

type ServerHandler func(ServerContext, protocol.Protocol)

var serverHandlerMgr map[protocol.ProtocolID]ServerHandler

func registerServerHandler() {
	serverHandlerMgr = make(map[protocol.ProtocolID]ServerHandler)
	serverHandlerMgr[msg.C2SMsgID_Login] = func(ctx ServerContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SLoginData)
		if c2sMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Login, p)
			return
		}

		user := ctx.GetUser(ctx.Link().UID())
		if user == nil {
			user = serverModel.NewUser()
			ctx.AddUser(ctx.Link().UID(), user)
		}

		s2cMsg := &msg.S2CLoginData{
			AccountID: int(ctx.Link().UID()),
			User:      user,
		}
		ctx.Send(event.New(msg.S2CMsgID_Login, s2cMsg))
	}
}

type UserContext interface {
	Send(*event.Event)
	User() *serverModel.User
}

type UserHandler func(UserContext, protocol.Protocol)

var userHandlerMgr map[protocol.ProtocolID]UserHandler

// business logic
// TODO: callback 使用 ctx 传递上下文，而不是直接传递 link？
// TODO: callback 直接传递 User 包裹 link？
func registerUserHandler() {
	userHandlerMgr = make(map[protocol.ProtocolID]UserHandler)
	userHandlerMgr[msg.C2SMsgID_Business] = func(ctx UserContext, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SBusinessData)
		if c2sMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_Business, p)
			return
		}

		ctx.User().AddCounter()

		s2cMsg := &msg.S2CBusinessData{
			Key:    c2sMsg.Key,
			Result: c2sMsg.Value1 + c2sMsg.Value2,
		}
		ctx.Send(event.New(msg.S2CMsgID_Business, s2cMsg))
	}
}

func (s *Server) serverHandlerMiddleware(l *link.Link, e *event.Event) {
	serverHandler, has := userHandlerMgr[e.ID()]
	if serverHandler == nil || !has {
		return
	}

	// s.userMgr

	// serverHandler()
}

func main() {
	// register protocol ID handler
	registerServerHandler() // use server context
	registerUserHandler()   // use user context

	// create server
	server := NewServer()

	// run server
	go server.Run()

	// watch system signal
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	server.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
