package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/example/sgs-server/server"
	"github.com/Mericusta/go-sgs/example/sgs-server/user"
)

// type UserContext interface {
// 	dispatcher.IContext
// 	User() *serverModel.User
// }

// // func NewUserContext() UserContext {
// // 	// return &
// // }

// type userHandlerMiddleware struct {
// 	f func(UserContext, protocol.Protocol)
// }

// func (m *userHandlerMiddleware) Do(l *link.Link, e *event.Event) {
// 	// userContext := newUser
// }

func main() {
	// register protocol ID handler
	server.RegisterHandler()   // use server context
	user.RegisterUserHandler() // use user context

	// create server
	sgs := server.NewServer()

	// append middleware
	sgs.AppendHandlerMiddleware(
		server.NewMiddleware(sgs),
		user.NewMiddleware(sgs),
	)

	// run server
	go sgs.Run()

	// watch system signal
	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt)
	<-s
	fmt.Printf("Note: close signal\n")
	close(s)
	fmt.Printf("Note: server exit\n")
	sgs.Exit() // end tcp listener, all link connection recv goroutine
	fmt.Printf("Note: waitting 5 seconds\n")
	time.Sleep(time.Second * 5)
}
