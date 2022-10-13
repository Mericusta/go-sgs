//go:build tlv || worker

package connector

import (
	"encoding/binary"
	"io"
	"robot-prototype/protocol"
)

func (c *MessageConnector) ReceiveMsg() (protocol.MSG_ID, []byte, error) {
	tagByteData := make([]byte, TLVPacketDataTagSize)
	_, readTagError := io.ReadFull(c.Connection, tagByteData)
	if readTagError != nil {
		return 0, nil, readTagError
	}
	tag := binary.BigEndian.Uint32(tagByteData)

	lengthByteData := make([]byte, TLVPacketDataLengthSize)
	_, readLengthError := io.ReadFull(c.Connection, lengthByteData)
	length := binary.BigEndian.Uint32(lengthByteData)
	if readLengthError != nil {
		return 0, nil, readLengthError
	}

	msgByteData := make([]byte, int(length))
	_, readMsgByteError := io.ReadFull(c.Connection, msgByteData)
	if readMsgByteError != nil {
		return 0, nil, readMsgByteError
	}
	return protocol.MSG_ID(tag), msgByteData, nil
}
