// go:build tlv || worker

package connector

import (
	"encoding/binary"
	"fmt"

	"github.com/Mericusta/go-sgs/msg"
)

func (c *MessageConnector) ReceiveMsg() (msg.MsgID, msg.Msg, error) {
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

	msg, err := msg.Unmarshal(msg.MsgID, msgValueByte)
	if err == nil {

	}

	return msg.MsgID(msgID), nil, nil
}
