package main

import (
	"context"
	"fmt"

	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/framework"
	"github.com/Mericusta/go-sgs/link"
	"github.com/Mericusta/go-sgs/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

// business logic
// TODO: callback 使用 ctx 传递上下文，而不是直接传递 link？
// TODO: callback 直接传递 Client 包裹 link？
var c2sMsgCallbackMap = make(map[protocol.ProtocolID]func(*link.Link, protocol.Protocol))

func registerC2SMsgCallback() {
	c2sMsgCallbackMap[C2SMsgID_HeartBeatCounter] = func(l *link.Link, c2sMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := c2sMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", C2SMsgID_HeartBeatCounter, c2sMsg)
			return
		}
		heartBeatCounterMsg.Count++
		l.Send(msg.New(S2CMsgID_HeartBeatCounter, heartBeatCounterMsg))
	}

	s2cMsgCallbackMap[S2CMsgID_HeartBeatCounter] = func(l *link.Link, s2cMsg protocol.Protocol) {
		heartBeatCounterMsg, ok := s2cMsg.(*HeartBeatCounter)
		if heartBeatCounterMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", C2SMsgID_HeartBeatCounter, s2cMsg)
			return
		}
		heartBeatCounterMsg.Count++
		l.Send(msg.New(C2SMsgID_HeartBeatCounter, heartBeatCounterMsg))
	}
}

func main() {
	// server
	registerC2SMsgCallback()
	serverCtx, serverCanceler := context.WithCancel(context.Background())
	server := framework.New(dispatcher.New(c2sMsgCallbackMap))
	go server.Run(serverCtx)
}
