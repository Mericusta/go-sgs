package framework

import (
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/Mericusta/go-sgs/acceptor"
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/dispatcher"
	link "github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type RunMiddleware interface {
	Do(dispatcher.IContext) bool
}

// Framework
type Framework struct {
	linkCounter         uint
	acceptorMgr         []acceptor.IAcceptor
	connChan            chan net.Conn
	dispatcherMgr       map[int]*dispatcher.Dispatcher
	handlerMgr          map[protocol.ProtocolID]dispatcher.FrameworkHandler
	handleMiddlewareMgr []dispatcher.HandlerMiddleware
	runMiddleware       RunMiddleware
}

func New() *Framework {
	return &Framework{
		connChan:      make(chan net.Conn, config.MaxConnectionCount),
		dispatcherMgr: make(map[int]*dispatcher.Dispatcher, config.DispatcherLinkerCount),
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
		logger.Log().Warn("framework not have any acceptor")
	}
}

func (f *Framework) singleRun() {
	a := f.acceptorMgr[0]
LOOP:
	for a.State() == acceptor.LISTENING {
		connection, acceptError := a.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				logger.Log().Info("framework single acceptor closed")
				break LOOP
			}
			logger.Log().Error("acceptor accept connection occurs error", zap.Error(acceptError))
			continue
		}
		f.run(connection)
	}
	logger.Log().Info("single acceptor end accept-goroutine")
}

func (f *Framework) accept(a acceptor.IAcceptor) {
LOOP:
	for a.State() == acceptor.LISTENING {
		connection, acceptError := a.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				logger.Log().Info("framework acceptor closed", zap.Int("acceptor", a.ID()))
				break LOOP
			}
			logger.Log().Error("acceptor accept connection occurs error", zap.Error(acceptError))
			continue
		}
		f.connChan <- connection
	}
	logger.Log().Info("acceptor end accept-goroutine", zap.Int("acceptor", a.ID()))
}

func (f *Framework) recvConn() {
	for connection := range f.connChan {
		f.run(connection)
	}
}

func (f *Framework) run(connection net.Conn) {
	l := link.New(connection)
	d, hasD := f.dispatch(l.UID())
	if err := d.Bind(l); err != nil {
		logger.Log().Error("dispatcher bind link occurs error", zap.Uint64("linker", l.UID()), zap.Int("dispatcher", d.Index()), zap.Error(err))
		return
	}
	logger.Log().Info("create link and bind dispatcher", zap.Uint64("linker", l.UID()), zap.Int("dispatcher", d.Index()), zap.Bool("newDispatcher", hasD))
	go l.HandleRecv()
	go l.HandleSend()
	if !hasD {
		d.SetHandleMiddleware(f.handleMiddlewareMgr)
		go d.HandleLogic()
	}
	if f.runMiddleware != nil {
		f.runMiddleware.Do(l)
	}
}

func (f *Framework) dispatch(uid uint64) (*dispatcher.Dispatcher, bool) {
	dispatcherIndex := int(uid) % (config.DispatcherCount)
	_, has := f.dispatcherMgr[dispatcherIndex]
	if !has {
		f.dispatcherMgr[int(dispatcherIndex)] = dispatcher.New(dispatcherIndex)
	}
	return f.dispatcherMgr[int(dispatcherIndex)], has
}

func (f *Framework) RegisterHandler(msgID protocol.ProtocolID, handler dispatcher.FrameworkHandler) {
	f.handlerMgr[msgID] = handler
}

func (f *Framework) AppendHandleMiddleware(hmd ...dispatcher.HandlerMiddleware) {
	f.handleMiddlewareMgr = append(f.handleMiddlewareMgr, hmd...)
}

func (f *Framework) SetRunMiddleware(rmd RunMiddleware) {
	f.runMiddleware = rmd
}

func (f *Framework) ForRangeDispatcher(handle func(uint64, *dispatcher.Dispatcher) bool) {
	for _, d := range f.dispatcherMgr {
		d.ForRangeLinker(handle)
	}
}

// Exit end acceptor, all link connection recv goroutine
func (f *Framework) Exit() {
	logger.Log().Info("close acceptor")
	var err error
	for _, acceptor := range f.acceptorMgr {
		err = acceptor.Close()
		if err != nil {
			logger.Log().Error("close acceptor occurs error", zap.Error(err))
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
	logger.Log().Info("stop signal")
	signal.Stop(s)
	close(s)
	logger.Log().Info("exit framework")
	f.Exit()
	logger.Log().Info("waitting 5 seconds")
	time.Sleep(time.Second * 5)
}
