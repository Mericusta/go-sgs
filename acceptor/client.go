package acceptor

import (
	"net"
	"time"
)

type ClientAcceptor struct {
	network         string
	addr            string
	tcpDialOvertime time.Duration
	state           AcceptorState
}

func NewClientAcceptor(network, addr string, tcpDialOvertime time.Duration) IAcceptor {
	return &ClientAcceptor{
		network:         network,
		addr:            addr,
		tcpDialOvertime: tcpDialOvertime,
		state:           LISTENING,
	}
}

func (a *ClientAcceptor) Accept() (net.Conn, error) {
	connection, dialError := net.DialTimeout(a.network, a.addr, a.tcpDialOvertime)
	if dialError != nil {
		return nil, dialError
	}
	a.state = CLOSED
	return connection, nil
}

func (a *ClientAcceptor) Close() error {
	a.state = CLOSED
	return nil
}

func (a *ClientAcceptor) State() AcceptorState {
	return a.state
}
