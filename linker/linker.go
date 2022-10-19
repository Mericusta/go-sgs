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
	ctx       context.Context // dispatcher make
}

func New(connection net.Conn) *Linker {
	return &Linker{
		connector: connector.New(connection),
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		recv:      make(chan *msg.Msg, config.ChannelBuffer),
		send:      make(chan *msg.Msg, config.ChannelBuffer),
		ctx:       context.Background(),
	}
}

func (l *Linker) UID() uint64 {
	return l.uid
}

func (l *Linker) Send(m *msg.Msg) {
	l.send <- m
}

func (l *Linker) Recv() (*msg.Msg, bool) {
	m, ok := <-l.recv
	return m, ok
}

// recv goroutine
func (l *Linker) HandleRecv() {
	for {
		protocolID, protocolData, err := l.connector.RecvMsg()
		if err != nil {
			fmt.Printf("Error: connector read tcp socket packet occurs error: %v\n", err.Error())
			if err != io.EOF {
				if opError, ok := err.(*net.OpError); ok && opError.Err != net.ErrClosed {
					// TODO: connector read error
					continue
				}
			}
			// tcp socket closed
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
		sendMsg, ok := <-l.send
		if !ok {
			fmt.Printf("Error: send msg is not ok\n")
			continue
		}
		err := l.connector.SendMsg(sendMsg.ID(), sendMsg.Data())
		if err != nil {
			// TODO: connector send error
			fmt.Printf("Error: connector send tcp socket packet occurs error: %v", err.Error())
		}
	}
}

// logic goroutine: 1 - 1 - 1
func (l *Linker) HandleLogic(handlerMap map[protocol.ProtocolID]func(*Linker, protocol.Protocol)) {
	for {
		select {
		case msg, ok := <-l.recv:
			if msg == nil || !ok {
				fmt.Printf("Error: linker logic goroutine receive msg is nil or not ok\n")
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
		case <-l.ctx.Done():
			fmt.Printf("Note: linker %v logic goroutine receive context done\n", l.uid)
			close(l.send)
			goto DONE
		}
	}
DONE:
	fmt.Printf("Note: linker %v logic goroutine done\n", l.uid)
}
