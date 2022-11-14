package dispatcher

import (
	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type FrameworkHandler func(IContext, protocol.Protocol)

type Dispatcher struct {
	l                   *link.Link
	eventChannel        chan *event.Event
	handlerMgr          map[protocol.ProtocolID]FrameworkHandler
	handleMiddlewareMgr []HandleMiddleware
	recoverMiddleware   RecoverMiddleware
}

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

func (d *Dispatcher) Dispatcher() *Dispatcher {
	return d
}

func (d *Dispatcher) SetHandleMiddleware(hmdMgr []HandleMiddleware) {
	d.handleMiddlewareMgr = hmdMgr
}

func (d *Dispatcher) SetRecoverMiddleware() {

}

func (d *Dispatcher) HandleLogic() {
	loopCounter := 0
	logger.Logger().Info("begin logic-goroutine", zap.Uint64("link", d.Link().UID()))
LOOP:
	for {
		logger.Logger().Debug("HandleLogic begin loop", zap.Int("loopCounter", loopCounter))
		loopCounter++
		select {
		case e, ok := <-d.eventChannel: // 主动发送，可以通过关闭 eventChannel 来退出，和 context 原理相同
			// 本地主动断开
			if !ok {
				logger.Logger().Info("event channel closed", zap.Uint64("link", d.Link().UID()))
				// d.Link().Exit() // 关闭 connection，退出发送协程
				// 关闭 connection 会导致接收协程退出
				break LOOP
				// TODO: 是否应该处理 recv 剩余的数据？
			}

			// 发送逻辑
			logger.Logger().Info("handle send-event", zap.Uint64("link", d.Link().UID()), zap.Any("event", e))
			// if d.handleIntercept(e) {
			// 	handler := d.handlerMgr[e.ID()]
			// 	if handler == nil {
			// 		fmt.Printf("Error: dispatcher event ID %v handler is nil", e.ID())
			// 		continue
			// 	}
			// 	handler(d, e.Data())
			// }
			d.Link().Send(e)
		case e, ok := <-d.Link().Recv(): // 被动接收
			// tcp 套接字已断开（远端/本地都有可能），recv 协程已退出
			// - 远端：需要关闭主动发送通道，需要退出发送协程，需要关闭 connection
			// - 本地：需要关闭主动发送通道，需要退出发送协程，不需要关闭 connection（重复关闭）
			// 	- 不可能由本地触发，因为 1-1-3 资源模型下，本地关闭只能由关闭 eventChannel 触发
			if !ok {
				logger.Logger().Info("recv-channel closed", zap.Uint64("link", d.Link().UID()))
				d.Link().Exit() // 关闭 connection
				logger.Logger().Info("close event-channel", zap.Uint64("link", d.Link().UID()))
				close(d.eventChannel) // 关闭主动发送通道
				break LOOP            // 退出逻辑协程
				// TODO: 是否需要处理 eventChannel 中剩余的内容？
			}

			// 接收逻辑
			logger.Logger().Info("handle recv-event", zap.Uint64("link", d.Link().UID()), zap.Any("event", e))
			d.handle(e)
			logger.Logger().Debug("handle done", zap.Uint64("link", d.Link().UID()))
		}
	}
	logger.Logger().Info("end logic-goroutine", zap.Uint64("link", d.Link().UID()))
}

func (d *Dispatcher) handle(e *event.Event) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logger.Logger().Warn("dispatcher handle panic info and recover\n", zap.Any("panicInfo", panicInfo))
		}
	}()

	if d.handleIntercept(e) {
		handler := d.handlerMgr[e.ID()]
		if handler == nil {
			logger.Logger().Error("dispatcher event ID handler is nil", zap.Uint32("ID", uint32(e.ID())))
			return
		}
		handler(d, e.Data())
	}
}

func (d *Dispatcher) handleIntercept(e *event.Event) bool {
	for _, handleMiddleware := range d.handleMiddlewareMgr {
		if !handleMiddleware.Do(d, e) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) Send(e *event.Event) {
	select {
	case d.eventChannel <- e:
	default:
		return
	}
}
