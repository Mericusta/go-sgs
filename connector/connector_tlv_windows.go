// go:build tlv || worker

package connector

import (
	"encoding/binary"
	"fmt"

	"github.com/Mericusta/go-sgs/protocol"
)

func (c *MessageConnector) RecvMsg() (protocol.ProtocolID, protocol.Protocol, error) {
	msgIDByte := make([]byte, TLVPacketDataTagSize)
	_, readTagError := c.Connection.Read(msgIDByte)
	if readTagError != nil {
		return 0, nil, readTagError
	}
	msgID := binary.BigEndian.Uint32(msgIDByte)

	msgLengthByte := make([]byte, TLVPacketDataLengthSize)
	_, readLengthError := c.Connection.Read(msgLengthByte)
	msgValueLength := binary.BigEndian.Uint32(msgLengthByte)
	if readLengthError != nil {
		return 0, nil, readLengthError
	}

	msgValueByte := make([]byte, int(msgValueLength))
	readLength, readMsgByteError := c.Connection.Read(msgValueByte)
	if readMsgByteError != nil {
		return 0, nil, readMsgByteError
	} else if readLength != int(msgValueLength) {
		return 0, nil, fmt.Errorf("read msg %v %v length %v not equal packet length %v", msgID, msgValueByte, readLength, msgValueLength)
	}

	msg, err := protocol.Unmarshal(protocol.ProtocolID(msgID), msgValueByte)
	if err != nil {
		return 0, nil, err
	} else if msg == nil {
		return 0, nil, fmt.Errorf("unmarshal msg %v %v got empty", msgID, msgValueByte)
	}

	return protocol.ProtocolID(msgID), nil, nil
}
