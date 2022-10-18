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

func NewLinker(connection net.Conn) *Linker {
	return &Linker{
		connector: connector.NewConnector(connection),
		uid:       uint64(time.Now().UnixNano()), // TODO: distributed-guid
		recv:      make(chan *msg.Msg, config.ChannelBuffer),
		send:      make(chan *msg.Msg, config.ChannelBuffer),
		ctx:       context.Background(),
	}
}

// recv goroutine
func (linker *Linker) HandleRecv() {
	for {
		protocolID, protocolData, err := linker.connector.RecvMsg()
		if err != nil {
			fmt.Printf("Error: connector read tcp socket packet occurs error: %v\n", err.Error())
			if err != io.EOF {
				if opError, ok := err.(*net.OpError); ok && opError.Err != net.ErrClosed {
					// TODO: connector read error
					continue
				}
			}
			// tcp socket closed
			close(linker.recv)
			linker.recv = nil
			return
		}
		linker.recv <- msg.NewMsg(protocolID, protocolData)
	}
}

// send goroutine
func (linker *Linker) HandleSend() {
	for {
		sendMsg, ok := <-linker.send
		if !ok {
			fmt.Printf("Error: send msg is not ok\n")
			continue
		}
		err := linker.connector.SendMsg(sendMsg.ID(), sendMsg.Msg())
		if err != nil {
			// TODO: connector send error
			fmt.Printf("Error: connector send tcp socket packet occurs error: %v", err.Error())
		}
	}
}
