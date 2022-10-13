//go:build tlv || worker

package connector

import (
	"encoding/binary"
	"io"
	"robot-prototype/protocol"
	"robot-prototype/ui"
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

type MessageConnector struct {
	BaseConnector
}

func (c *MessageConnector) SendMsg(msgID protocol.MSG_ID, msgByteData []byte) error {
	msgByteDataLength := uint32(len(msgByteData))
	tlvPackMsg := make([]byte, TLVPacketDataTagSize+TLVPacketDataLengthSize+msgByteDataLength)

	// tlvPackMsg[0,TLVPacketDataTagSize]
	binary.BigEndian.PutUint32(tlvPackMsg, uint32(msgID))

	// tlvPackMsg[TLVPacketDataTagSize,TLVPacketDataTagSize+TLVPacketDataLengthSize]
	binary.BigEndian.PutUint32(tlvPackMsg[TLVPacketDataTagSize:], msgByteDataLength)

	// tlvPackMsg[TLVPacketDataTagSize+TLVPacketDataLengthSize:]
	copy(tlvPackMsg[TLVPacketDataTagSize+TLVPacketDataLengthSize:], msgByteData)

	ui.OutputDebugInfo("tlv pack msg [%v, %v] byte data = %v", msgID, msgByteData, tlvPackMsg)

	_, writeError := c.Connection.Write(tlvPackMsg)
	if writeError != nil {
		if writeError != io.EOF {
			ui.OutputErrorInfo("connection write occurs error: %v", writeError)
		}
		return writeError
	}

	return nil
}
