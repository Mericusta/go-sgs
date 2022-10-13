//go:build tbp

package connector

import (
	"encoding/binary"
	"robot-prototype/global"
	"robot-prototype/logger"
	"robot-prototype/protocol"
	"robot-prototype/utility"
)

// ┌──────────────────────────────────┬──────────────────┐
// │              Header              │       Body       │
// ├──────┬─────────────┬─────────────┼───────┬──────────┤
// │ Type │ Body Length │ Packet Flag │ MsgID │ Msg Data │
// │──────┼─────────────┼─────────────┼───────┼──────────┤
// │  1   │      4      │       1     │   4   │          │
// └──────┴─────────────┴─────────────┴───────┴──────────┘

const (
	// TBPPacketTypeLength TBP 数据包 type 数据长度
	TBPPacketTypeSize = 1

	// TBPBodyLengthSize TBP 数据包 body_length 数据长度
	TBPBodyLengthSize = 4

	// TBPPacketFlagSize TBP 数据包 packet_flag 数据长度
	TBPPacketFlagSize = 1

	// TBPHeaderSize TBP 数据包头数据长度
	TBPHeaderSize = TBPPacketTypeSize + TBPBodyLengthSize + TBPPacketFlagSize

	// TBPMsgIDSize TBP 数据包 msg_id 数据长度
	TBPMsgIDSize = 4

	// TBPPacketType TBP 数据包 packet_type 数据
	TBPPacketType = 'P'

	// TBPPacketType TBP 数据包 心跳 packet_type 数据
	TBPHeartBeatPacketType = 'T'

	// TBPPacketFlag TBP 数据包 packet_flag 数据
	TBPPacketFlag = ' '
)

type MessageConnector struct {
	BaseConnector
}

func (c *MessageConnector) SendMsg(msgID protocol.MSG_ID, msgByteData []byte) error {
	var tbpPackMsg []byte
	if msgID == protocol.MSG_ID_C2S_HEART_BEAT {
		tbpPackMsg = make([]byte, TBPHeaderSize)
		// packet type
		tbpPackMsg[0] = TBPHeartBeatPacketType
	} else {
		dir, id := utility.ResolveMsgID(uint32(msgID))
		logger.Log(global.LogModuleTBPConnector, logger.LogLevelDebug, "send msg dir %v, id %v", dir, id)

		msgByteDataLength := uint32(len(msgByteData))
		TBPBodySize := TBPMsgIDSize + msgByteDataLength
		tbpPackMsg = make([]byte, TBPHeaderSize+TBPBodySize)
		// packet type
		tbpPackMsg[0] = TBPPacketType
		// body length
		binary.BigEndian.PutUint32(tbpPackMsg[TBPPacketTypeSize:], TBPBodySize)
		// packet flag
		tbpPackMsg[TBPPacketTypeSize+TBPBodyLengthSize] = TBPPacketFlag
		// msg id
		binary.BigEndian.PutUint32(tbpPackMsg[TBPHeaderSize:], uint32(msgID))
		// msg data
		copy(tbpPackMsg[TBPHeaderSize+TBPMsgIDSize:], msgByteData)
	}

	_, writeError := c.Connection.Write(tbpPackMsg)
	return writeError
}
