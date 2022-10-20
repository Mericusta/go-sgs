package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/linker"
	"github.com/Mericusta/go-sgs/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

var s2cMsgCallbackMap = make(map[protocol.ProtocolID]func(*linker.Linker, protocol.Protocol))

func registerS2CMsgCallback() {
	s2cMsgCallbackMap[S2CMsgID_HeartBeatCounter] = func(linker *linker.Linker, s2cMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := s2cMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", C2SMsgID_HeartBeatCounter, s2cMsg)
			return
		}
		heartBeatCounterMsg.Count++
		linker.Send(msg.New(C2SMsgID_HeartBeatCounter, heartBeatCounterMsg))
	}
}

type Client struct {
	*linker.Linker
	counter int
}

func NewClient(l *linker.Linker) *Client {
	return &Client{Linker: l}
}
