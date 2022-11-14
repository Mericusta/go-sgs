package link

import (
	"io"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/logger"
	"github.com/Mericusta/go-sgs/packer"
	"go.uber.org/zap"
)

type LINK_STATE int

const (
	LINK_INIT LINK_STATE = iota
	LINK_CONNECTED
	LINK_CLOSED
)

type Link struct {
	uid    uint64
	state  LINK_STATE // TODO: 有并发问题
	packer packer.Packer
	recv   chan *event.Event // TODO: 不要传递小对象
	send   chan *event.Event // TODO: 不要传递小对象
}

func New(connection net.Conn) *Link {
	return &Link{
		uid:    uint64(time.Now().UnixNano()), // TODO: distributed-guid
		state:  LINK_CONNECTED,
		packer: packer.New(connection),
		recv:   make(chan *event.Event, config.ChannelBuffer),
		send:   make(chan *event.Event, config.ChannelBuffer),
	}
}

func (l *Link) UID() uint64 {
	return l.uid
}

func (l *Link) Send(m *event.Event) {
	if m == nil || l.state != LINK_CONNECTED {
		logger.Logger().Debug("link send nil or state is not LINK_CONNECTED", zap.Uint64("link", l.UID()), zap.Bool("isNil", m == nil), zap.Int("state", int(l.state)))
		return
	}
	// TODO: 通过长度判断一下是否可以 send，以免在 send-channel 缓存满了并且被关闭之后阻塞在这里
	// logger.Logger().Debug("link send-channel length", zap.Uint64("link", l.UID()), zap.Int("length", len(l.send)))
	l.send <- m
}

func (l *Link) Recv() <-chan *event.Event {
	return l.recv
}

// recv goroutine
func (l *Link) HandleRecv() {
	logger.Logger().Info("begin recv goroutine", zap.Uint64("link", l.UID()))
LOOP:
	for {
		protocolID, protocolData, err := l.packer.RecvMsg()
		if err != nil {
			if err == io.EOF {
				logger.Logger().Info("tcp socket closed by remote", zap.Uint64("link", l.UID()))
			} else if opError, ok := err.(*net.OpError); ok && opError.Err == net.ErrClosed {
				logger.Logger().Info("tcp socket closed by local", zap.Uint64("link", l.UID()))
			} else {
				logger.Logger().Error("tcp socket read packet occurs error", zap.Uint64("link", l.UID()), zap.Error(err))
			}
			logger.Logger().Info("close recv-channel", zap.Uint64("link", l.UID()))
			close(l.recv)
			break LOOP
		} else {
			l.recv <- event.New(protocolID, protocolData)
		}
	}
	logger.Logger().Info("end recv-goroutine", zap.Uint64("link", l.UID()))
}

// send goroutine
func (l *Link) HandleSend() {
	logger.Logger().Info("begin send-goroutine", zap.Uint64("link", l.UID()))
LOOP:
	for {
		sendMsg, ok := <-l.send
		if !ok {
			logger.Logger().Info("send-channel closed", zap.Uint64("link", l.UID()))
			break LOOP
		}
		err := l.packer.SendMsg(sendMsg.ID(), sendMsg.Data())
		if err != nil {
			logger.Logger().Error("send tcp socket packet occurs error", zap.Uint64("link", l.UID()), zap.Error(err))
			if err == io.EOF {
				logger.Logger().Info("tcp socket occurs io.EOF", zap.Uint64("link", l.UID()))
				break LOOP
			}
			continue
		}
	}
	logger.Logger().Info("end send-goroutine", zap.Uint64("link", l.UID()))
}

// // logic goroutine
// // TODO: handle logic 不一定只由 link.recv 来触发，handle logic 本身是可以由数据驱动的（比如每隔一段时间主动推送消息）
// func (l *Link) HandleLogic(ctx context.Context, handlerMap map[protocol.ProtocolID]func(*Link, protocol.Protocol)) {
// 	for {
// 		select {
// 		case e, ok := <-l.recv:
// 			// close 的时候会触发 e == nil && ok == false，此时代表已关闭，需要 return
// 			// 但是结束逻辑会由 context 的 cancel 提前触发，所以此处一般用不到
// 			if e == nil || !ok {
// 				fmt.Printf("Error: link %v logic goroutine receive event is nil %v or not ok %v", l.uid, e == nil, ok)
// 				continue
// 			}

// 			// TODO: how to get callback without creating callback map for every link ?
// 			// - use global value map, multi goroutine read concurrently, also can not write after register
// 			// - use sync.Map, but mutex is performance bottle neck
// 			callback := handlerMap[e.ID()]
// 			if callback == nil {
// 				fmt.Printf("Error: event ID %v callback is nil", e.ID())
// 				continue
// 			}

// 			// TODO: make context
// 			callback(l, e.Data())
// 		case <-ctx.Done():
// 			fmt.Printf("Note: link %v receive context done and end logic goroutine", l.uid)
// 			l.Exit()
// 			goto DONE
// 		}
// 	}
// DONE:
// 	fmt.Printf("Note: link %v logic goroutine done", l.uid)
// }

// exit tcp socket
func (l *Link) Exit() {
	if l.state == LINK_CLOSED {
		logger.Logger().Info("link already exit", zap.Uint64("link", l.uid))
		return
	}
	logger.Logger().Info("link exit", zap.Uint64("link", l.uid))
	// 标记状态，防止逻辑协程在 handler 中可能会往已关闭的 channel 中发送数据从而导致阻塞
	l.state = LINK_CLOSED
	// 主动断开 tcp socket
	logger.Logger().Info("close packer", zap.Uint64("link", l.uid))
	err := l.packer.Close()
	if err != nil {
		logger.Logger().Warn("close packer occurs error", zap.Error(err))
	}
	// 退出 send 协程
	logger.Logger().Info("close send-channel", zap.Uint64("link", l.uid))
	close(l.send)
}
