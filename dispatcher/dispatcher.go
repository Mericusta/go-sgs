package dispatcher

import (
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

type Dispatcher struct {
	handlerMgr map[protocol.ProtocolID]func(*link.Link, protocol.Protocol)
}

// TODO: 使用 dispatcher 传入 Link 的方式导致 client 和 server 不能复用 dispatcher
// 因为 client 对每一个 link 的用户层抽象是 Client
// 而 server 对每一个 link 的用户层抽象是 User
// 可以考虑使用泛型或者接口来进行统一抽象
func New(handlerMgr map[protocol.ProtocolID]func(*link.Link, protocol.Protocol)) *Dispatcher {
	return &Dispatcher{
		handlerMgr: handlerMgr,
	}
}

func (d *Dispatcher) HandlerMap() map[protocol.ProtocolID]func(*link.Link, protocol.Protocol) {
	return d.handlerMgr
}
