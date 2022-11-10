package acceptor

import "net"

type AcceptorState uint8

const (
	IDLE AcceptorState = iota + 1
	LISTENING
	CLOSED
)

type IAcceptor interface {
	ID() int
	Accept() (net.Conn, error)
	Close() error
	State() AcceptorState // TODO: 有并发问题
}

type BaseAcceptor struct {
	id int
}

func NewBaseAcceptor(id int) *BaseAcceptor {
	return &BaseAcceptor{id: id}
}

func (a *BaseAcceptor) ID() int {
	return a.id
}
