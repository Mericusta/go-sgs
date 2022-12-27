package dispatcher

import (
	"fmt"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/protocol"
	"go.uber.org/zap"
)

type FrameworkHandler func(IContext, protocol.ProtocolMsg)

type Dispatcher struct {
	index                int
	linkerMap            map[uint64]*linker.Linker
	eventChannel         chan *event.Event
	recvChannel          chan *event.Event
	handlerMgr           map[protocol.ProtocolID]FrameworkHandler
	handlerMiddlewareMgr []HandlerMiddleware
	recoverMiddleware    RecoverMiddleware
}

func New(index int) *Dispatcher {
	return &Dispatcher{
		index:        index,
		linkerMap:    make(map[uint64]*linker.Linker, config.DispatcherLinkerCount),
		eventChannel: make(chan *event.Event, config.ChannelBuffer),
		recvChannel:  make(chan *event.Event, config.ChannelBuffer),
		handlerMgr:   make(map[protocol.ProtocolID]FrameworkHandler),
	}
}

func (d *Dispatcher) Index() int {
	return d.index
}

func (d *Dispatcher) Bind(l *linker.Linker) error {
	if _, has := d.linkerMap[l.UID()]; has {
		return fmt.Errorf("link %v already exists in the dispatcher %v", l.UID(), d.index)
	}
	d.linkerMap[l.UID()] = l
	l.Bind(d.recvChannel)
	return nil
}

func (d *Dispatcher) Linker(UID uint64) *linker.Linker {
	return d.linkerMap[UID]
}

func (d *Dispatcher) Dispatcher() *Dispatcher {
	return d
}

func (d *Dispatcher) SetHandleMiddleware(hmdMgr []HandlerMiddleware) {
	d.handlerMiddlewareMgr = hmdMgr
}

func (d *Dispatcher) SetRecoverMiddleware() {

}

func (d *Dispatcher) HandleLogic() {
	loopCounter := 0
LOOP:
	for {
		logger.Log().Debug("HandleLogic begin loop", zap.Int("dispatcher", d.index), zap.Int("loopCounter", loopCounter))
		loopCounter++
		select {
		case e, ok := <-d.eventChannel: // 主动发送，可以通过关闭 eventChannel 来退出，和 context 原理相同
			if !ok {
				logger.Log().Info("event channel closed", zap.Int("dispatcher", d.index))
				// 关闭 connection 会导致接收协程退出
				break LOOP
			}

			// 发送逻辑
			logger.Log().Info("handle send-event", zap.Uint64("linker", e.LinkerUID()))
			d.Linker(e.LinkerUID()).Send(e)
		case e, ok := <-d.recvChannel: // 被动接收
			if !ok {
				logger.Log().Info("recv-channel closed", zap.Int("dispatcher", d.index))
				logger.Log().Info("close event-channel", zap.Int("dispatcher", d.index))
				close(d.eventChannel) // 关闭主动发送通道
				break LOOP            // 退出逻辑协程
			}

			// 接收逻辑
			logger.Log().Info("handle recv-event", zap.Int("dispatcher", d.index), zap.Uint64("linker", e.LinkerUID()))
			d.handle(e)
			logger.Log().Debug("handle done", zap.Int("dispatcher", d.index), zap.Uint64("linker", e.LinkerUID()))
		}
	}
	logger.Log().Info("end logic-goroutine", zap.Int("dispatcher", d.index))
}

func (d *Dispatcher) handle(e *event.Event) {
	defer func() {
		if panicInfo := recover(); panicInfo != nil {
			logger.Log().Warn("dispatcher handle panic info and recover\n", zap.Any("panicInfo", panicInfo))
		}
	}()

	if d.handleIntercept(e) {
		handler := d.handlerMgr[e.ID()]
		if handler == nil {
			logger.Log().Error("dispatcher event ID handler is nil", zap.Uint32("ID", uint32(e.ID())))
			return
		}
		handler(d.Linker(e.LinkerUID()), e.Data())
	}
}

func (d *Dispatcher) handleIntercept(e *event.Event) bool {
	for _, handleMiddleware := range d.handlerMiddlewareMgr {
		if !handleMiddleware.Do(d.Linker(e.LinkerUID()), e) {
			return false
		}
	}
	return true
}

func (d *Dispatcher) Send(e *event.Event) {
	select {
	case d.eventChannel <- e:
		logger.Log().Info("send event to dispatcher linker", zap.Int("dispatcher", d.index), zap.Uint64("linker", e.LinkerUID()))
	default:
		return
	}
}

func (d *Dispatcher) ForRangeLinker(handle func(uint64, *Dispatcher) bool) {
	for uid := range d.linkerMap {
		handle(uid, d)
	}
}
