package moduleNet

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ProtocolMsg any
type ProtocolName protoreflect.Name

var protocolMakerMap map[ProtocolName]func() ProtocolMsg = make(map[ProtocolName]func() ProtocolMsg)

func RegisterProtocolMaker(id ProtocolName, f func() ProtocolMsg) {
	protocolMakerMap[id] = f
}

func NewProtocolMsg(id ProtocolName) (ProtocolMsg, error) {
	maker := protocolMakerMap[id]
	if maker == nil {
		return nil, fmt.Errorf("unknown protocol id %v", id)
	}
	p := maker()
	if p == nil {
		return nil, fmt.Errorf("protocol id %v maker make a nil protocol", id)
	}
	return p, nil
}

func RegisterHandler[T any, RT proto.Message, C any](m map[ProtocolName]func(C, proto.Message, string) (proto.Message, error), handler func(C, T, string) (RT, error)) ProtocolName {
	msgID := ProtocolName(func(t any) proto.Message { return t.(proto.Message) }(*new(T)).ProtoReflect().Descriptor().Name())
	f := func(ctx C, iMsg proto.Message, uuid string) (proto.Message, error) {
		if msg, ok := iMsg.(T); ok {
			res, err := handler(ctx, msg, uuid)
			return res, err
		} else {
			return nil, fmt.Errorf("msgID %v data %v type assert failed", msgID, msg)
		}
	}
	m[msgID] = f
	return msgID
}
