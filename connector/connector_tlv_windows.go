//go:build tlv

package connector

import (
	"encoding/binary"
	"fmt"

	"github.com/Mericusta/go-sgs/protocol"
)

func (c *MessageConnector) RecvMsg() (protocol.ProtocolID, protocol.Protocol, error) {
	tagBytes := make([]byte, TLVPacketDataTagSize)
	_, readTagError := c.Connection.Read(tagBytes)
	if readTagError != nil {
		return 0, nil, readTagError
	}
	tag := binary.BigEndian.Uint32(tagBytes)

	lengthBytes := make([]byte, TLVPacketDataLengthSize)
	_, readLengthError := c.Connection.Read(lengthBytes)
	if readLengthError != nil {
		return 0, nil, readLengthError
	}
	length := binary.BigEndian.Uint32(lengthBytes)

	valueBytes := make([]byte, int(length))
	readValueLength, readValueError := c.Connection.Read(valueBytes)
	if readValueError != nil {
		return 0, nil, readValueError
	} else if readValueLength != int(length) {
		return 0, nil, fmt.Errorf("read msg %v %v length %v not equal packet length %v", tag, valueBytes, readValueLength, length)
	}

	msg, err := protocol.Unmarshal(protocol.ProtocolID(tag), valueBytes)
	if err != nil {
		return 0, nil, err
	} else if msg == nil {
		return 0, nil, fmt.Errorf("unmarshal msg %v %v got empty", tag, valueBytes)
	}

	return protocol.ProtocolID(tag), msg, nil
}
