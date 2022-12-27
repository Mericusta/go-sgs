//go:build tlv

package packer

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/Mericusta/go-sgs/protocol"
)

func (c *MessagePacker) Unpack() (protocol.ProtocolID, protocol.ProtocolMsg, error) {
	tagBytes := make([]byte, TLVPacketDataTagSize)
	_, readTagError := io.ReadFull(c.Connection, tagBytes)
	if readTagError != nil {
		return 0, nil, readTagError
	}
	tag := binary.BigEndian.Uint32(tagBytes)

	lengthBytes := make([]byte, TLVPacketDataLengthSize)
	_, readLengthError := io.ReadFull(c.Connection, lengthBytes)
	if readLengthError != nil {
		return 0, nil, readLengthError
	}
	length := binary.BigEndian.Uint32(lengthBytes)

	valueBytes := make([]byte, int(length))
	_, readValueError := io.ReadFull(c.Connection, valueBytes)
	if readValueError != nil {
		return 0, nil, readValueError
	}

	msg, err := protocol.Unmarshal(protocol.ProtocolID(tag), valueBytes)
	if err != nil {
		return 0, nil, err
	} else if msg == nil {
		return 0, nil, fmt.Errorf("unmarshal msg %v %v got empty", tag, valueBytes)
	}

	return protocol.ProtocolID(tag), msg, nil
}
