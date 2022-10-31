package acceptor

import "net"

type IAcceptor interface {
	Accept() (net.Conn, error)
	Close() error
}
