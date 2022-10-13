//go:build tbp

package connector

import (
	"encoding/binary"
	"io"
	"robot-prototype/global"
	"robot-prototype/logger"
	"robot-prototype/protocol"
	"robot-prototype/utility"
)

func (c *MessageConnector) ReceiveMsg() (protocol.MSG_ID, []byte, error) {
	msgTypeByteData := make([]byte, TBPPacketTypeSize)
	_, readTagError := io.ReadFull(c.Connection, msgTypeByteData)
	if readTagError != nil {
		return 0, nil, readTagError
	}

	bodyLengthByteData := make([]byte, TBPBodyLengthSize)
	_, readBodyLengthError := io.ReadFull(c.Connection, bodyLengthByteData)
	if readBodyLengthError != nil {
		return 0, nil, readBodyLengthError
	}
	bodyLength := binary.BigEndian.Uint32(bodyLengthByteData)

	packetFlagByteData := make([]byte, TBPPacketFlagSize)
	_, readPacketFlagError := io.ReadFull(c.Connection, packetFlagByteData)
	if readPacketFlagError != nil {
		return 0, nil, readPacketFlagError
	}

	if msgTypeByteData[0] == TBPHeartBeatPacketType {
		return protocol.MSG_ID_S2C_HEART_BEAT, nil, nil
	}

	xIDByteData := make([]byte, TBPMsgIDSize)
	_, readXIDError := io.ReadFull(c.Connection, xIDByteData)
	if readXIDError != nil {
		return 0, nil, readXIDError
	}
	xID := binary.BigEndian.Uint32(xIDByteData)

	msgByteData := make([]byte, int(bodyLength-TBPMsgIDSize))
	_, readMsgDataError := io.ReadFull(c.Connection, msgByteData)
	if readMsgDataError != nil {
		return 0, nil, readMsgDataError
	}

	dir, id := utility.ResolveMsgID(xID)
	logger.Log(global.LogModuleTBPConnector, logger.LogLevelDebug, "recv msg dir %v, id %v", dir, id)

	return protocol.MSG_ID(xID), msgByteData, nil
}
