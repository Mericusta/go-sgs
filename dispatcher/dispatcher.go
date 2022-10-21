package dispatcher

import (
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

type Dispatcher struct {
	handlerMgr map[protocol.ProtocolID]func(*link.Link, protocol.Protocol)
}

func New(handlerMgr map[protocol.ProtocolID]func(*link.Link, protocol.Protocol)) *Dispatcher {
	return &Dispatcher{
		handlerMgr: handlerMgr,
	}
}

func (d *Dispatcher) HandlerMap() map[protocol.ProtocolID]func(*link.Link, protocol.Protocol) {
	return d.handlerMgr
}
