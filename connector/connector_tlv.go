//go:build tlv

package connector

import (
	"encoding/binary"
	"fmt"

	"github.com/Mericusta/go-sgs/protocol"
)

// ┌─────┬────────┬───────┐
// │ Tag │ Length │ Value │
// ├─────┼────────┼───────┤
// │  4  │   4    │       │
// └─────┴────────┴───────┘

const (
	// TLV 格式数据包中数据的标识的值的占位长度
	TLVPacketDataTagSize = 4

	// TLV 格式数据包中数据的长度的值的占位长度
	TLVPacketDataLengthSize = 4
)

// 根据编译选项来确定 connector，所以名称必须一致
type MessageConnector struct {
	BaseConnector
}

func (c *MessageConnector) SendMsg(msgID protocol.ProtocolID, msgData protocol.Protocol) error {
	msgValueByte, err := protocol.Marshal(msgData)
	if len(msgValueByte) == 0 {
		return fmt.Errorf("marshal msg %v %v got empty slice", msgID, msgData)
	}
	if err != nil {
		return err
	}

	msgByteDataLength := len(msgValueByte)
	tlvPacketLength := TLVPacketDataTagSize + TLVPacketDataLengthSize + msgByteDataLength
	tlvPacket := make([]byte, tlvPacketLength)

	// tlvPackMsg[0,TLVPacketDataTagSize]
	binary.BigEndian.PutUint32(tlvPacket, uint32(msgID))

	// tlvPackMsg[TLVPacketDataTagSize,TLVPacketDataTagSize+TLVPacketDataLengthSize]
	binary.BigEndian.PutUint32(tlvPacket[TLVPacketDataTagSize:], uint32(msgByteDataLength))

	// tlvPackMsg[TLVPacketDataTagSize+TLVPacketDataLengthSize:]
	copy(tlvPacket[TLVPacketDataTagSize+TLVPacketDataLengthSize:], msgValueByte)

	writeLength, writeError := c.BaseConnector.Connection.Write(tlvPacket)
	if writeError != nil {
		return writeError
	} else if writeLength != tlvPacketLength {
		return fmt.Errorf("write msg %v %v length %v not equal packet length %v", msgID, msgData, writeLength, msgByteDataLength)
	}

	return writeError
}
