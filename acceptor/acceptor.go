package acceptor

import "net"

type AcceptorState uint8

const (
	IDLE AcceptorState = iota + 1
	LISTENING
	CLOSED
)

type IAcceptor interface {
	Accept() (net.Conn, error)
	Close() error
	State() AcceptorState // TODO: 有并发问题
}
