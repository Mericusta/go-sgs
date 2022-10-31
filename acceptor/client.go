package acceptor

import "net"

type ClientAcceptor struct {
}

func NewClientAcceptor(concurrent bool) IAcceptor {
	return &ClientAcceptor{}
}

func (a *ClientAcceptor) Accept() (net.Conn, error) {
	return nil, nil
}

func (a *ClientAcceptor) Close() error {
	return nil
}
