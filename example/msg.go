package main

import "github.com/Mericusta/go-sgs/protocol"

const (
	C2SMsgID_HeartBeatCounter = iota + 1
	S2CMsgID_HeartBeatCounter
)

type HeartBeatCounter struct {
	Count int `json:"count"`
}

func init() {
	protocol.RegisterMsgMaker(
		protocol.ProtocolID(C2SMsgID_HeartBeatCounter), func() any { return &HeartBeatCounter{} },
	)
}
