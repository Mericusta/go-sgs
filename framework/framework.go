package framework

import (
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-logger"
	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

type RunMiddleware interface {
	Do(dispatcher.IContext) bool
}

// Framework
type Framework struct {
	linkCounter         uint
	acceptorMgr         []acceptor.IAcceptor
	connChan            chan net.Conn
	handleMiddlewareMgr []dispatcher.HandleMiddleware
	dispatcherMgr       map[uint64]*dispatcher.Dispatcher
	handlerMgr          map[protocol.ProtocolID]dispatcher.FrameworkHandler
	runMiddleware       RunMiddleware
}

func New() *Framework {
	return &Framework{
		connChan:      make(chan net.Conn, config.MaxConnectionCount),
		dispatcherMgr: make(map[uint64]*dispatcher.Dispatcher, config.MaxConnectionCount),
		handlerMgr:    make(map[protocol.ProtocolID]dispatcher.FrameworkHandler),
	}
}

func (f *Framework) AppendAcceptor(acceptor acceptor.IAcceptor) {
	f.acceptorMgr = append(f.acceptorMgr, acceptor)
}

func (f *Framework) Run() {
	acceptorCount := len(f.acceptorMgr)
	switch {
	case acceptorCount > 1:
		go f.recvConn()
		for _, acceptor := range f.acceptorMgr {
			go f.accept(acceptor)
		}
	case acceptorCount == 1:
		f.singleRun()
	default:
		logger.Warn().Package("framework").Content("not have any acceptor")
	}
}

func (f *Framework) singleRun() {
	a := f.acceptorMgr[0]
	for a.State() == acceptor.LISTENING {
		connection, acceptError := a.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				logger.Info().Package("framework").Package("Framework").Func("singleRun").Content("acceptor closed")
				return
			}
			logger.Error().Package("framework").Package("Framework").Func("singleRun").Content("acceptor accept connection occurs error: %v", acceptError.Error())
			continue
		}
		f.run(connection)
	}
}

func (f *Framework) accept(a acceptor.IAcceptor) {
	for a.State() == acceptor.LISTENING {
		connection, acceptError := a.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				logger.Info().Package("framework").Package("Framework").Func("accept").Content("acceptor closed")
				return
			}
			logger.Error().Package("framework").Package("Framework").Func("accept").Content("acceptor accept connection occurs error: %v", acceptError.Error())
			continue
		}
		f.connChan <- connection
	}
}

func (f *Framework) recvConn() {
	for connection := range f.connChan {
		f.run(connection)
	}
}

func (f *Framework) run(connection net.Conn) {
	l := link.New(connection)
	d := dispatcher.New(l)
	f.dispatcherMgr[l.UID()] = d
	logger.Info().Package("framework").Package("Framework").Func("run").Content("create link %v and its dispatcher", l.UID())
	d.SetHandleMiddleware(f.handleMiddlewareMgr)
	logger.Info().Package("framework").Package("Framework").Func("run").Content("link %v begin recv goroutine", l.UID())
	go l.HandleRecv()
	logger.Info().Package("framework").Package("Framework").Func("run").Content("link %v begin send goroutine", l.UID())
	go l.HandleSend()
	logger.Info().Package("framework").Package("Framework").Func("run").Content("link %v begin logic goroutine", l.UID())
	go d.HandleLogic()
	if f.runMiddleware != nil {
		f.runMiddleware.Do(d)
	}
}

func (f *Framework) RegisterHandler(msgID protocol.ProtocolID, handler dispatcher.FrameworkHandler) {
	f.handlerMgr[msgID] = handler
}

func (f *Framework) AppendHandleMiddleware(hmd ...dispatcher.HandleMiddleware) {
	f.handleMiddlewareMgr = append(f.handleMiddlewareMgr, hmd...)
}

func (f *Framework) SetRunMiddleware(rmd RunMiddleware) {
	f.runMiddleware = rmd
}

func (f *Framework) ForRangeDispatcher(handle func(uint64, *dispatcher.Dispatcher) bool) {
	for id, d := range f.dispatcherMgr {
		handle(id, d)
	}
}

// Exit end acceptor, all link connection recv goroutine
func (f *Framework) Exit() {
	logger.Info().Package("framework").Package("Framework").Func("Exit").Content("close acceptor")
	var err error
	for _, acceptor := range f.acceptorMgr {
		err = acceptor.Close()
		if err != nil {
			logger.Error().Package("framework").Package("Framework").Func("Exit").Content("close acceptor occurs error: %v", err.Error())
		}
	}
	// 只需要退出 dispatcher，dispatcher 退出会引起
	// for _, l := range s.linkMgr {
	// 	fmt.Printf("Note: server close link %v connection", l.UID())
	// }
}

func (f *Framework) Hold() {
	s := make(chan os.Signal, 10)
	signal.Notify(s, os.Interrupt)
	<-s
	logger.Info().Package("framework").Package("Framework").Func("Hold").Content("stop signal")
	signal.Stop(s)
	close(s)
	logger.Info().Package("framework").Package("Framework").Func("Hold").Content("exit framework")
	f.Exit()
	logger.Info().Package("framework").Package("Framework").Func("Hold").Content("waitting 5 seconds")
	time.Sleep(time.Second * 5)
}
