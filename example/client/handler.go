package main

import (
	"fmt"

	"github.com/Mericusta/go-sgs/example/common"
	"github.com/Mericusta/go-sgs/example/msg"
	"github.com/Mericusta/go-sgs/protocol"
)

var clientMsgCallbackMap = make(map[protocol.ProtocolID]func(*common.Client, protocol.Protocol))

func registerClientMsgCallback() {
	clientMsgCallbackMap[msg.S2CMsgID_CalculatorAdd] = func(client *common.Client, p protocol.Protocol) {
		s2cMsg, ok := p.(*msg.S2CCalculatorData)
		if s2cMsg == nil || !ok {
			fmt.Printf("Error: msg ID %v data %+v not match\n", msg.S2CMsgID_CalculatorAdd, p)
			return
		}
		if client.GetExpect(s2cMsg.Key) != s2cMsg.Result {
			fmt.Printf("Error: client %v S2CMsgID_CalculatorAdd key %v s2c msg result %v not match expect %v\n", client.Index(), s2cMsg.Key, s2cMsg.Result, client.GetExpect(s2cMsg.Key))
			return
		}
		client.DelExpect(s2cMsg.Key)
	}
}
