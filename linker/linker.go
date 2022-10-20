package linker

import (
	"context"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/Mericusta/go-sgs/config"
	"github.com/Mericusta/go-sgs/connector"
	"github.com/Mericusta/go-sgs/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

// Linker
// SELECT 1: 相同 linker，在经过不同编译条件的情况下，可以处理不同格式的 packet
// NO-NEED 2: 相同 linker，在不经过编译的情况下，可以处理不同格式的 packet
// - 在没有“额外信息（如何处理 packet）”的情况下不知道 packet 的处理格式
// 	- 提前发包告知
// 	- 每个包内告知

type Linker struct {
	uid       uint64
	connector connector.Connector
	recv      chan *msg.Msg
	send      chan *msg.Msg
	// ctx       context.Context // dispatcher make
}

func New(connection net.Conn) *Linker {
	return &Linker{
		connector: connector.New(connection),
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		recv:      make(chan *msg.Msg, config.ChannelBuffer),
		send:      make(chan *msg.Msg, config.ChannelBuffer),
		// ctx:       context.Background(),
	}
}

func (l *Linker) UID() uint64 {
	return l.uid
}

func (l *Linker) Send(m *msg.Msg) {
	if m == nil {
		return
	}
	l.send <- m
}

func (l *Linker) Recv() <-chan *msg.Msg {
	return l.recv
}

// recv goroutine
func (l *Linker) HandleRecv() {
	for {
		protocolID, protocolData, err := l.connector.RecvMsg()
		if err != nil {
			if err != io.EOF {
				if opError, ok := err.(*net.OpError); ok && opError.Err != net.ErrClosed {
					// TODO: connector read error
					fmt.Printf("Error: linker %v read tcp socket packet occurs error: %v\n", l.uid, err.Error())
					continue
				}
			}
			// tcp socket closed
			fmt.Printf("Note: linker %v tcp socket is closed by remote, then close recv channel and end recv goroutine\n", l.uid)
			close(l.recv)
			l.recv = nil
			return
		}
		l.recv <- msg.New(protocolID, protocolData)
	}
}

// send goroutine
func (l *Linker) HandleSend() {
	for {
		sendMsg, ok := <-l.send // TODO: close 的时候会触发 msg == nil && ok == false，此时代表已关闭，需要 return
		if sendMsg == nil || !ok {
			// fmt.Printf("Error: linker %v send goroutine msg is nil %v or not ok %v\n", l.uid, sendMsg == nil, ok)
			fmt.Printf("Error: linker %v send goroutine msg is nil %v or not ok %v, end send goroutine\n", l.uid, sendMsg == nil, ok)
			return
		}
		err := l.connector.SendMsg(sendMsg.ID(), sendMsg.Data())
		if err != nil {
			fmt.Printf("Error: linker %v send tcp socket packet occurs error: %v", l.uid, err.Error())
			if err == io.EOF {
				// TODO: connector send error
				fmt.Printf("Note: linker %v tcp socket occurs io.EOF and end send goroutine\n", l.uid)
				return
			}
			continue
		}
	}
}

// logic goroutine
func (l *Linker) HandleLogic(ctx context.Context, handlerMap map[protocol.ProtocolID]func(*Linker, protocol.Protocol)) {
	for {
		select {
		case msg, ok := <-l.recv:
			// close 的时候会触发 msg == nil && ok == false，此时代表已关闭，需要 return
			// 但是结束逻辑会由 context 的 cancel 提前触发，所以此处一般用不到
			if msg == nil || !ok {
				fmt.Printf("Error: linker %v logic goroutine receive msg is nil %v or not ok %v\n", l.uid, msg == nil, ok)
				continue
			}

			// TODO: how to get callback without creating callback map for every linker ?
			// - use global value map, multi goroutine read concurrently, also can not write after register
			// - use sync.Map, but mutex is performance bottle neck
			callback := handlerMap[msg.ID()]
			if callback == nil {
				fmt.Printf("Error: msg ID %v callback is nil\n", msg.ID())
				continue
			}

			// TODO: make context
			callback(l, msg.Data())
		case <-ctx.Done():
			fmt.Printf("Note: linker %v receive context done and end logic goroutine\n", l.uid)
			close(l.send)
			goto DONE
		}
	}
DONE:
	fmt.Printf("Note: linker %v logic goroutine done\n", l.uid)
}

func (l *Linker) Close() {
	l.connector.Close()
}
