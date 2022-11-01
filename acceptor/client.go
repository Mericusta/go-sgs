package acceptor

import (
	"net"
	"time"
)

type ClientAcceptor struct {
	network         string
	addr            string
	tcpDialOvertime time.Duration
}

func NewClientAcceptor(network, addr string, tcpDialOvertime time.Duration) IAcceptor {
	return &ClientAcceptor{
		network:         network,
		addr:            addr,
		tcpDialOvertime: tcpDialOvertime,
	}
}

func (a *ClientAcceptor) Accept() (net.Conn, error) {
	connection, dialError := net.DialTimeout(a.network, a.addr, a.tcpDialOvertime)
	if dialError != nil {
		return nil, dialError
	}
	return connection, &net.OpError{Err: net.ErrClosed}
}

func (a *ClientAcceptor) Close() error {
	return nil
}
