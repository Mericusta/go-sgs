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

// Link
// SELECT 1: 相同 linker，在经过不同编译条件的情况下，可以处理不同格式的 packet
// NO-NEED 2: 相同 linker，在不经过编译的情况下，可以处理不同格式的 packet
// - 在没有“额外信息（如何处理 packet）”的情况下不知道 packet 的处理格式
// 	- 提前发包告知
// 	- 每个包内告知

type Link struct {
	uid       uint64
	connector connector.Connector
	recv      chan *event.Event // TODO: 不要传递小对象
	send      chan *event.Event // TODO: 不要传递小对象
	// ctx       context.Context // dispatcher make
}

func New(connection net.Conn) *Link {
	return &Link{
		connector: connector.New(connection),
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		recv:      make(chan *event.Event, config.ChannelBuffer),
		send:      make(chan *event.Event, config.ChannelBuffer),
		// ctx:       context.Background(),
	}
}

func (l *Link) UID() uint64 {
	return l.uid
}

func (l *Link) Send(m *event.Event) {
	if m == nil {
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
			if err != io.EOF {
				if opError, ok := err.(*net.OpError); ok && opError.Err != net.ErrClosed {
					// TODO: connector read error
					fmt.Printf("Error: link %v read tcp socket packet occurs error: %v\n", l.uid, err.Error())
					continue
				}
			}
			// tcp socket closed
			fmt.Printf("Note: link %v tcp socket is closed by remote, then close recv channel and end recv goroutine\n", l.uid)
			close(l.recv)
			return
		}
		l.recv <- event.New(protocolID, protocolData)
	}
}

// send goroutine
func (l *Link) HandleSend() {
	for {
		sendMsg, ok := <-l.send // TODO: connector close 的时候会触发 event == nil && ok == false，此时代表已关闭，需要 return
		if sendMsg == nil || !ok {
			// fmt.Printf("Error: link %v send goroutine event is nil %v or not ok %v\n", l.uid, sendMsg == nil, ok)
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

// 主动断开 tcp socket
func (l *Link) Close() {
	l.connector.Close()
}

// 被动断开 tcp socket
func (l *Link) Exit() {
	close(l.send)
}
