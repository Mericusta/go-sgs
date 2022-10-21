package framework

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/link"
)

// Framework
type Framework struct {
	listener   net.Listener
	linkMgr    []*link.Link
	dispatcher *dispatcher.Dispatcher
}

func New(dispatcher *dispatcher.Dispatcher) *Framework {
	var listener net.Listener
	var listenError error
	if config.TcpKeepAliveSeconds > 0 {
		listenCfg := net.ListenConfig{KeepAlive: time.Second * 5}
		listener, listenError = listenCfg.Listen(context.Background(), "tcp", config.DefaultServerAddress)
	} else {
		listener, listenError = net.Listen("tcp", config.DefaultServerAddress)
	}
	if listener == nil || listenError != nil {
		fmt.Printf("Error: listen tcp %v occurs error: %v\n", config.DefaultServerAddress, listenError.Error())
		return nil
	}

	return &Framework{
		listener:   listener,
		linkMgr:    make([]*link.Link, 0, config.MaxConnectionCount),
		dispatcher: dispatcher,
	}
}

func (s *Framework) Run(ctx context.Context) {
	for {
		connection, acceptError := s.listener.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				fmt.Printf("Note: server listener closed\n")
				return
			}
			fmt.Printf("Error: server listener accept connection occurs error: %v\n", acceptError.Error())
			continue
		}

		l := link.New(connection)
		go l.HandleRecv()
		go l.HandleSend()
		go l.HandleLogic(ctx, s.dispatcher.HandlerMap()) // TODO: dispatcher
		s.linkMgr = append(s.linkMgr, l)
		fmt.Printf("Note: server create link %v\n", l.UID())
	}
}

func (s *Framework) Exit() {
	fmt.Printf("Note: server close listener\n")
	s.listener.Close()
	for _, l := range s.linkMgr {
		fmt.Printf("Note: server close link %v connection\n", l.UID())
		l.Close()
	}
}
