//go:build protobuf

package protocol

import (
	"google.golang.org/protobuf/proto"
)

func Marshal(v ProtocolMsg) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func Unmarshal(id ProtocolID, b []byte) (ProtocolMsg, error) {
	msg, err := newProtocolMsg(id)
	if msg == nil || err != nil {
		return nil, err
	}
	err = proto.Unmarshal(b, msg.(proto.Message))
	return msg, err
}
