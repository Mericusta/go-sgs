package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/example/common"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

var serverMsgCallbackMap = make(map[protocol.ProtocolID]func(*common.User, protocol.Protocol))

// business logic
// TODO: callback 使用 ctx 传递上下文，而不是直接传递 link？
// TODO: callback 直接传递 User 包裹 link？
func registerServerMsgCallback() {
	serverMsgCallbackMap[msg.C2SMsgID_CalculatorAdd] = func(user *common.User, p protocol.Protocol) {
		c2sMsg, ok := p.(*msg.C2SCalculatorData)
		if c2sMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.C2SMsgID_CalculatorAdd, p)
			return
		}
		user.AddCounter()
		s2cMsg := &msg.S2CCalculatorData{
			Key:    c2sMsg.Key,
			Result: c2sMsg.Value1 + c2sMsg.Value2,
		}
		user.Send(event.New(msg.S2CMsgID_CalculatorAdd, s2cMsg))
	}
}
