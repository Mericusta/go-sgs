package dispatcher

import (
	"fmt"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/middleware"
	"github.com/Mericusta/go-sgs/protocol"
)

type FrameworkHandler func(middleware.IContext, protocol.Protocol)

type Dispatcher struct {
	l                    *link.Link
	eventChannel         chan *event.Event
	handlerMgr           map[protocol.ProtocolID]FrameworkHandler
	handlerMiddlewareMgr []middleware.HandlerMiddleware
}

// TODO: 使用 dispatcher 传入 Link 的方式导致 client 和 server 不能复用 dispatcher
// 因为 client 对每一个 link 的用户层抽象是 Client
// 而 server 对每一个 link 的用户层抽象是 User
// 可以考虑使用泛型或者接口来进行统一抽象
func New(l *link.Link) *Dispatcher {
	return &Dispatcher{
		l:            l,
		eventChannel: make(chan *event.Event, config.ChannelBuffer),
		handlerMgr:   make(map[protocol.ProtocolID]FrameworkHandler),
	}
}

func (d *Dispatcher) Link() *link.Link {
	return d.l
}

func (d *Dispatcher) SetHandlerMiddleware(hmdMgr []middleware.HandlerMiddleware) {
	d.handlerMiddlewareMgr = hmdMgr
}

func (d *Dispatcher) HandleLogic() {
LOOP:
	for {
		select {
		case e, ok := <-d.eventChannel: // 主动发送，可以通过关闭 eventChannel 来退出，和 context 原理相同
			// 本地主动断开
			if !ok {
				fmt.Printf("Note: dispatcher link %v event channel closed\n", d.Link().UID())
				d.Link().Exit() // 关闭 connector，退出发送协程
				// 关闭 connector 会导致接收协程退出
				break LOOP
				// TODO: 是否应该处理 recv 剩余的数据？
			}

			// 发送逻辑
			fmt.Printf("Note: dispatcher link %v handle send logic, event %+v\n", d.Link().UID(), e)
			if d.handlerIntercept(e) {
				handler := d.handlerMgr[e.ID()]
				if handler == nil {
					fmt.Printf("Error: dispatcher event ID %v handler is nil\n", e.ID())
					continue
				}
				handler(d, e.Data())
			}
		case e, ok := <-d.Link().Recv(): // 被动接收
			// tcp 套接字已断开（远端/本地都有可能），recv 协程已退出
			// - 远端：需要关闭主动发送通道，需要退出发送协程，需要关闭 connector
			// - 本地：需要关闭主动发送通道，需要退出发送协程，不需要关闭 connector（重复关闭）
			// 	- 不可能由本地触发，因为 1-1-3 资源模型下，本地关闭只能由关闭 eventChannel 触发
			if !ok {
				fmt.Printf("Note: dispatcher link %v receive channel closed\n", d.Link().UID())
				d.Link().Exit()       // 关闭 connector，退出发送协程
				close(d.eventChannel) // 关闭主动发送通道
				break LOOP            // 退出逻辑协程
				// TODO: 是否需要处理 eventChannel 中剩余的内容？
			}

			// 接收逻辑
			fmt.Printf("Note: dispatcher link %v handle recv logic, event %+v\n", d.Link().UID(), e)
			if d.handlerIntercept(e) {
				handler := d.handlerMgr[e.ID()]
				if handler == nil {
					fmt.Printf("Error: dispatcher event ID %v handler is nil\n", e.ID())
					continue
				}
				handler(d, e.Data())
			}
		}
	}
}

func (d *Dispatcher) handlerIntercept(e *event.Event) bool {
	for _, handlerMiddleware := range d.handlerMiddlewareMgr {
		if !handlerMiddleware.Do(d, e) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) Exit() {
	close(d.eventChannel)
}
