package dispatcher

import (
	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/protocol"
)

type Dispatcher struct {
	handlerMgr map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol)
}

func New(handlerMgr map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol)) *Dispatcher {
	return &Dispatcher{
		handlerMgr: handlerMgr,
	}
}

func (d *Dispatcher) HandlerMap() map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol) {
	return d.handlerMgr
}
