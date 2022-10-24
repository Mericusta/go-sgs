package dispatcher

import (
	"context"
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/protocol"
)

type Dispatcher struct {
	eventChannel chan *event.Event
	handlerMgr   map[protocol.ProtocolID]func(*link.Link, protocol.Protocol)
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

func (d *Dispatcher) HandleLogic(ctx context.Context, link *link.Link) {
LOOP:
	for {
		select {
		case e, ok := <-d.eventChannel: // 主动发送
		priority:
			for {
				select {
				case <-ctx.Done(): // 主动结束，本端主动通过 context cancel 结束，必须保证本端先关闭 event channel
					fmt.Printf("Note: dispatcher link %v context done\n", link.UID())
					goto LOOP
				default:
					break priority
				}
			}
			// 发送逻辑
			fmt.Printf("Note: dispatcher link %v handle send logic\n", link.UID())
			if !ok { // 由于对端断开 tcp 套接字而结束
				fmt.Printf("Note: dispatcher link %v send channel closed\n", link.UID())
				break LOOP
			}
			handler := d.handlerMgr[e.ID()]
			if handler == nil {
				fmt.Printf("Error: dispatcher event ID %v handler is nil\n", e.ID())
				continue
			}
			handler(link, e.Data())
		case e, ok := <-link.Recv(): // 被动接收
			if !ok { // 被动结束，对端断开 tcp 套接字
				fmt.Printf("Note: dispatcher link %v receive channel closed\n", link.UID())
				close(d.eventChannel)
				goto LOOP
			}
			// 接收逻辑
			fmt.Printf("Note: dispatcher link %v handle recv logic\n", link.UID())
			handler := d.handlerMgr[e.ID()]
			if handler == nil {
				fmt.Printf("Error: dispatcher event ID %v handler is nil\n", e.ID())
				continue
			}
			handler(link, e.Data())
		}
	}
}

func (d *Dispatcher) Exit() {
	close(d.eventChannel)
}
