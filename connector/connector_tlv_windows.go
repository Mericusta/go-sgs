// go:build tlv || worker

package connector

import (
	"encoding/binary"

	"github.com/Mericusta/go-sgs/msg"
)

func (c *MessageConnector) ReceiveMsg() (msg.MsgID, []byte, error) {
	tagByteData := make([]byte, TLVPacketDataTagSize)
	_, readTagError := c.Connection.Read(tagByteData)
	if readTagError != nil {
		return 0, nil, readTagError
	}
	tag := binary.BigEndian.Uint32(tagByteData)

	lengthByteData := make([]byte, TLVPacketDataLengthSize)
	_, readLengthError := c.Connection.Read(lengthByteData)
	length := binary.BigEndian.Uint32(lengthByteData)
	if readLengthError != nil {
		return 0, nil, readLengthError
	}

	msgByteData := make([]byte, int(length))
	_, readMsgByteError := c.Connection.Read(msgByteData)
	if readMsgByteError != nil {
		return 0, nil, readMsgByteError
	}
	return msg.MsgID(tag), msgByteData, nil
}
