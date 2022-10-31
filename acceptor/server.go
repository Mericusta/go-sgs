package acceptor

import (
	"context"
	"fmt"
	"net"
	"time"
)

type ServerAcceptor struct {
	net.Listener
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
		fmt.Printf("Error: listen tcp %v occurs error: %v\n", addr, listenError.Error())
		return nil
	}
	return &ServerAcceptor{Listener: listener}
}
