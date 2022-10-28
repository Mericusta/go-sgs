package framework

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

// Framework
type Framework struct {
	listener   net.Listener
	linkMgr    map[uint64]*link.Link
	handlerMgr map[protocol.ProtocolID]dispatcher.FrameworkHandler
	dispatcher map[uint64]*dispatcher.Dispatcher
}

func New() *Framework {
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
		linkMgr:    make(map[uint64]*link.Link, config.MaxConnectionCount),
		dispatcher: make(map[uint64]*dispatcher.Dispatcher, config.MaxConnectionCount),
	}
}

// func (s *Framework) Run(ctx context.Context) {
// 	for {
// 		connection, acceptError := s.listener.Accept()
// 		if acceptError != nil {
// 			if acceptError.(*net.OpError).Err == net.ErrClosed {
// 				fmt.Printf("Note: server listener closed\n")
// 				return
// 			}
// 			fmt.Printf("Error: server listener accept connection occurs error: %v\n", acceptError.Error())
// 			continue
// 		}

// 		l := link.New(connection)
// 		go l.HandleRecv()
// 		go l.HandleSend()
// 		go l.HandleLogic(ctx, s.dispatcher.HandlerMap()) // TODO: dispatcher
// 		s.linkMgr = append(s.linkMgr, l)
// 		fmt.Printf("Note: server create link %v\n", l.UID())
// 	}
// }

// TODO: 客户端 vs 服务器，相同的 framework
// - 暴露问题1：handle logic 必须定义在 framework 中
// - 暴露问题2：handle logic 的 link 的包裹体，必须成为 framework 中的一部分，否则就得用接口
func (s *Framework) Run() {
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

		// TODO: 可以考虑在这里启 go 协程，吧 linkMgr 和 dispatcher 改成 sync.Map
		l := link.New(connection)
		s.linkMgr[l.UID()] = l
		d := dispatcher.New()
		s.dispatcher[l.UID()] = d
		fmt.Printf("Note: server create link and dispatcher %v\n", l.UID())
		fmt.Printf("Note: link begin recv goroutine %v\n", l.UID())
		go l.HandleRecv()
		fmt.Printf("Note: link begin send goroutine %v\n", l.UID())
		go l.HandleSend()
		fmt.Printf("Note: dispatcher begin logic goroutine %v\n", l.UID())
		go d.HandleLogic(l)
	}
}

func (s *Framework) RegisterHandler(msgID protocol.ProtocolID, handler dispatcher.FrameworkHandler) {

}

func (s *Framework) Exit() {
	fmt.Printf("Note: server close listener\n")
	s.listener.Close()
	// 只需要退出 dispatcher，dispatcher 退出会引起
	// for _, l := range s.linkMgr {
	// 	fmt.Printf("Note: server close link %v connection\n", l.UID())
	// }
}
