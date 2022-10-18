//go:build tlv || protobuf

package protocol

import (
	"google.golang.org/protobuf/proto"
)

func Marshal(v any) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func Unmarshal(id ProtocolID, b []byte) (any, error) {
	msg, err := newMsg(id)
	if msg == nil || err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, msg.(proto.Message))
	return msg, err
}
