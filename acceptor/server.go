package acceptor

import (
	"context"
	"net"
	"time"

	"github.com/Mericusta/go-logger"
)

type ServerAcceptor struct {
	net.Listener
	state AcceptorState
}

func NewServerAcceptor(network, addr string, tcpKeepAlive time.Duration) IAcceptor {
	var listener net.Listener
	var listenError error
	if tcpKeepAlive > 0 {
		listenCfg := net.ListenConfig{KeepAlive: tcpKeepAlive}
		listener, listenError = listenCfg.Listen(context.Background(), network, addr)
	} else {
		listener, listenError = net.Listen(network, addr)
	}
	if listener == nil || listenError != nil {
		logger.Error().Package("acceptor").Func("NewServerAcceptor").Content("listen tcp %v occurs error: %v", addr, listenError.Error())
		return nil
	}
	return &ServerAcceptor{Listener: listener, state: LISTENING}
}

func (a *ServerAcceptor) Close() error {
	a.state = CLOSED
	return a.Listener.Close()
}

func (a *ServerAcceptor) State() AcceptorState {
	return a.state
}
