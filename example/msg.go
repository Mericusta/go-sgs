package main

import "github.com/Mericusta/go-sgs/protocol"

const (
	MsgIDHeartBeatCounter = iota + 1
)

type HeartBeatCounter struct {
	Count int `json:"count"`
}

func init() {
	protocol.RegisterMsgMaker(
		protocol.ProtocolID(MsgIDHeartBeatCounter), func() any { return &HeartBeatCounter{} },
	)
}
