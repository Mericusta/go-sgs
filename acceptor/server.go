package acceptor

import (
	"context"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/logger"
	"go.uber.org/zap"
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
		logger.Logger().Error("listen tcp addr occurs error", zap.String("addr", addr), zap.Error(listenError))
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
