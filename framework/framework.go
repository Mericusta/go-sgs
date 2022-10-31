package framework

import (
	"fmt"
	"net"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/middleware"
	"github.com/Mericusta/go-sgs/protocol"
)

// Framework
type Framework struct {
	acceptorMgr         []acceptor.IAcceptor
	connChan            chan net.Conn
	handlerMgr          map[protocol.ProtocolID]dispatcher.FrameworkHandler
	handleMiddlewareMgr []middleware.HandleMiddleware
	dispatcherMgr       map[uint64]*dispatcher.Dispatcher
}

func New() *Framework {
	return &Framework{
		acceptorMgr: []acceptor.IAcceptor{
			acceptor.NewServerAcceptor("tcp", config.DefaultServerAddress, config.TcpKeepAlive),
		},
		connChan:      make(chan net.Conn, config.MaxConnectionCount),
		dispatcherMgr: make(map[uint64]*dispatcher.Dispatcher, config.MaxConnectionCount),
	}
}

func (s *Framework) Run() {
	for {
		connection, acceptError := s.acceptor.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				fmt.Printf("Note: server listener closed\n")
				return
			}
			fmt.Printf("Error: server listener accept connection occurs error: %v\n", acceptError.Error())
			continue
		}
		l := link.New(connection)
		d := dispatcher.New(l)
		s.dispatcherMgr[l.UID()] = d
		d.SetHandleMiddleware(s.handleMiddlewareMgr)
		fmt.Printf("Note: server create link and dispatcher %v\n", l.UID())
		fmt.Printf("Note: link begin recv goroutine %v\n", l.UID())
		go l.HandleRecv()
		fmt.Printf("Note: link begin send goroutine %v\n", l.UID())
		go l.HandleSend()
		fmt.Printf("Note: dispatcher begin logic goroutine %v\n", l.UID())
		go d.HandleLogic()
	}
}

func (s *Framework) RegisterHandler(msgID protocol.ProtocolID, handler dispatcher.FrameworkHandler) {

}

func (s *Framework) AppendHandleMiddleware(hmd ...middleware.HandleMiddleware) {
	s.handleMiddlewareMgr = append(s.handleMiddlewareMgr, hmd...)
}

func (s *Framework) Exit() {
	fmt.Printf("Note: server close acceptor\n")
	err := s.acceptor.Close()
	fmt.Printf("Error: server close acceptor occurs error: %v\n", err)
	// 只需要退出 dispatcher，dispatcher 退出会引起
	// for _, l := range s.linkMgr {
	// 	fmt.Printf("Note: server close link %v connection\n", l.UID())
	// }
}
