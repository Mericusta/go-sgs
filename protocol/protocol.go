package protocol

import "fmt"

type Protocol any
type ProtocolID uint32

var msgMap map[ProtocolID]func() any = make(map[ProtocolID]func() any)

func RegisterMsgMaker(id ProtocolID, f func() any) {
	msgMap[id] = f
}

func newMsg(id ProtocolID) (any, error) {
	msgMaker := msgMap[id]
	if msgMaker == nil {
		return nil, fmt.Errorf("unknown msg id %v", id)
	}
	msg := msgMaker()
	if msg == nil {
		return nil, fmt.Errorf("msg id %v maker make nil msg", id)
	}
	return msg, nil
}
