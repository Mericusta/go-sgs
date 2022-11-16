package linker

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

type Linker struct {
	uid    uint64
	state  LINK_STATE        // TODO: 有并发问题
	packer packer.Packer     // 通过
	recv   chan *event.Event // recv-channel TODO: 不要传递小对象
	send   chan *event.Event // send-channel TODO: 不要传递小对象
}

func New(conn net.Conn) *Linker {
	return &Linker{
		uid:    uint64(time.Now().UnixNano()), // TODO: distributed-guid
		state:  LINK_CONNECTED,
		packer: packer.New(conn),
		recv:   make(chan *event.Event, config.ChannelBuffer),
		send:   make(chan *event.Event, config.ChannelBuffer),
	}
}

func (l *Linker) UID() uint64 {
	return l.uid
}

func (l *Linker) Send(m *event.Event) {
	if m == nil || l.state != LINK_CONNECTED {
		logger.Logger().Debug("linker send nil or state is not LINK_CONNECTED", zap.Uint64("uid", l.uid), zap.Bool("isNil", m == nil), zap.Int("state", int(l.state)))
		return
	}
	// TODO: 通过长度判断一下是否可以 send，以免在 send-channel 缓存满了并且被关闭之后阻塞在这里
	// logger.Logger().Debug("linker send-channel length", zap.Uint64("uid", l.uid), zap.Int("length", len(l.send)))
	l.send <- m
}

func (l *Linker) Recv() <-chan *event.Event {
	return l.recv
}

// recv-goroutine
func (l *Linker) HandleRecv() {
	logger.Logger().Info("begin recv goroutine", zap.Uint64("uid", l.uid))
LOOP:
	for {
		protocolID, protocolData, err := l.packer.Unpack()
		if err != nil {
			if err == io.EOF {
				logger.Logger().Info("tcp socket closed by remote", zap.Uint64("uid", l.uid))
			} else if opError, ok := err.(*net.OpError); ok && opError.Err == net.ErrClosed {
				logger.Logger().Info("tcp socket closed by local", zap.Uint64("uid", l.uid))
			} else {
				logger.Logger().Error("tcp socket read packet occurs error", zap.Uint64("uid", l.uid), zap.Error(err))
			}
			logger.Logger().Info("close recv-channel", zap.Uint64("uid", l.uid))
			close(l.recv)
			break LOOP
		} else {
			l.recv <- event.New(protocolID, protocolData)
		}
	}
	logger.Logger().Info("end recv-goroutine", zap.Uint64("uid", l.uid))
}

// send-goroutine
func (l *Linker) HandleSend() {
	logger.Logger().Info("begin send-goroutine", zap.Uint64("uid", l.uid))
LOOP:
	for {
		sendMsg, ok := <-l.send
		if !ok {
			logger.Logger().Info("send-channel closed", zap.Uint64("uid", l.uid))
			break LOOP
		}
		err := l.packer.Pack(sendMsg.ID(), sendMsg.Data())
		if err != nil {
			logger.Logger().Error("send tcp socket packet occurs error", zap.Uint64("uid", l.uid), zap.Error(err))
			if err == io.EOF {
				logger.Logger().Info("tcp socket occurs io.EOF", zap.Uint64("uid", l.uid))
				break LOOP
			}
			continue
		}
	}
	logger.Logger().Info("end send-goroutine", zap.Uint64("uid", l.uid))
}

// exit tcp socket
func (l *Linker) Exit() {
	if l.state == LINK_CLOSED {
		logger.Logger().Info("linker already exit", zap.Uint64("uid", l.uid))
		return
	}
	logger.Logger().Info("linker exit", zap.Uint64("uid", l.uid))
	// 标记状态，防止逻辑协程在 handler 中可能会往已关闭的 channel 中发送数据从而导致阻塞
	l.state = LINK_CLOSED
	// 主动断开 tcp socket
	logger.Logger().Info("close packer", zap.Uint64("uid", l.uid))
	err := l.packer.Close()
	if err != nil {
		logger.Logger().Warn("close packer occurs error", zap.Error(err))
	}
	// 退出 send 协程
	logger.Logger().Info("close send-channel", zap.Uint64("uid", l.uid))
	close(l.send)
}
