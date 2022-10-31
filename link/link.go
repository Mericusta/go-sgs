package link

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/connector"
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/protocol"
)

type LINK_STATE int

const (
	LINK_INIT LINK_STATE = iota
	LINK_CONNECTED
	LINK_CLOSED
)

type Link struct {
	uid       uint64
	state     LINK_STATE
	connector connector.Connector
	recv      chan *event.Event // TODO: 不要传递小对象
	send      chan *event.Event // TODO: 不要传递小对象
}

func New(connection net.Conn) *Link {
	return &Link{
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		state:     LINK_CONNECTED,
		connector: connector.New(connection),
		recv:      make(chan *event.Event, config.ChannelBuffer),
		send:      make(chan *event.Event, config.ChannelBuffer),
	}
}

func (l *Link) UID() uint64 {
	return l.uid
}

func (l *Link) Send(m *event.Event) {
	if m == nil || l.state != LINK_CONNECTED {
		return
	}
	l.send <- m
}

func (l *Link) Recv() <-chan *event.Event {
	return l.recv
}

// recv goroutine
func (l *Link) HandleRecv() {
	for {
		protocolID, protocolData, err := l.connector.RecvMsg()
		if err != nil {
			if err == io.EOF {
				fmt.Printf("Note: link %v tcp socket closed by remote\n", l.uid)
			} else if opError, ok := err.(*net.OpError); ok && opError.Err == net.ErrClosed {
				fmt.Printf("Note: link %v tcp socket closed by local\n", l.uid)
			} else {
				fmt.Printf("Note: link %v tcp socket read packet occurs error: %v\n", l.uid, err.Error())
				continue
			}
			fmt.Printf("Note: link %v close recv channel\n", l.uid)
			close(l.recv)
			return
		} else {
			l.recv <- event.New(protocolID, protocolData)
		}
	}
}

// send goroutine
func (l *Link) HandleSend() {
	for {
		sendMsg, ok := <-l.send // TODO: connector close 的时候会触发 event == nil && ok == false，此时代表已关闭，需要 return
		if sendMsg == nil || !ok {
			fmt.Printf("Error: link %v send goroutine event is nil %v or not ok %v, end send goroutine\n", l.uid, sendMsg == nil, ok)
			return
		}
		err := l.connector.SendMsg(sendMsg.ID(), sendMsg.Data())
		if err != nil {
			fmt.Printf("Error: link %v send tcp socket packet occurs error: %v", l.uid, err.Error())
			if err == io.EOF {
				// TODO: connector send error
				fmt.Printf("Note: link %v tcp socket occurs io.EOF and end send goroutine\n", l.uid)
				return
			}
			continue
		}
	}
}

// logic goroutine
// TODO: handle logic 不一定只由 link.recv 来触发，handle logic 本身是可以由数据驱动的（比如每隔一段时间主动推送消息）
func (l *Link) HandleLogic(ctx context.Context, handlerMap map[protocol.ProtocolID]func(*Link, protocol.Protocol)) {
	for {
		select {
		case e, ok := <-l.recv:
			// close 的时候会触发 e == nil && ok == false，此时代表已关闭，需要 return
			// 但是结束逻辑会由 context 的 cancel 提前触发，所以此处一般用不到
			if e == nil || !ok {
				fmt.Printf("Error: link %v logic goroutine receive event is nil %v or not ok %v\n", l.uid, e == nil, ok)
				continue
			}

			// TODO: how to get callback without creating callback map for every link ?
			// - use global value map, multi goroutine read concurrently, also can not write after register
			// - use sync.Map, but mutex is performance bottle neck
			callback := handlerMap[e.ID()]
			if callback == nil {
				fmt.Printf("Error: event ID %v callback is nil\n", e.ID())
				continue
			}

			// TODO: make context
			callback(l, e.Data())
		case <-ctx.Done():
			fmt.Printf("Note: link %v receive context done and end logic goroutine\n", l.uid)
			l.Exit()
			goto DONE
		}
	}
DONE:
	fmt.Printf("Note: link %v logic goroutine done\n", l.uid)
}

// tcp socket
func (l *Link) Exit() {
	// 标记状态，防止逻辑协程在 handler 中可能会往已关闭的 channel 中发送数据从而导致阻塞
	l.state = LINK_CLOSED
	// 主动断开 tcp socket
	l.connector.Close()
	// 退出 send 协程
	close(l.send)
}
