//go:build tbp

package connector

import (
	"encoding/binary"
	"robot-prototype/global"
	"robot-prototype/logger"
	"robot-prototype/protocol"
	"robot-prototype/utility"
)

func (c *MessageConnector) ReceiveMsg() (protocol.MSG_ID, []byte, error) {
	msgTypeByteData := make([]byte, TBPPacketTypeSize)
	_, readTagError := c.Connection.Read(msgTypeByteData)
	if readTagError != nil {
		return 0, nil, readTagError
	}

	bodyLengthByteData := make([]byte, TBPBodyLengthSize)
	_, readBodyLengthError := c.Connection.Read(bodyLengthByteData)
	if readBodyLengthError != nil {
		return 0, nil, readBodyLengthError
	}
	bodyLength := binary.BigEndian.Uint32(bodyLengthByteData)

	packetFlagByteData := make([]byte, TBPPacketFlagSize)
	_, readPacketFlagError := c.Connection.Read(packetFlagByteData)
	if readPacketFlagError != nil {
		return 0, nil, readPacketFlagError
	}

	if msgTypeByteData[0] == TBPHeartBeatPacketType {
		return protocol.MSG_ID_S2C_HEART_BEAT, nil, nil
	}

	xIDByteData := make([]byte, TBPMsgIDSize)
	_, readXIDError := c.Connection.Read(xIDByteData)
	if readXIDError != nil {
		return 0, nil, readXIDError
	}
	xID := binary.BigEndian.Uint32(xIDByteData)

	msgByteData := make([]byte, int(bodyLength-TBPMsgIDSize))
	_, readMsgDataError := c.Connection.Read(msgByteData)
	if readMsgDataError != nil {
		return 0, nil, readMsgDataError
	}

	dir, id := utility.ResolveMsgID(xID)
	logger.Log(global.LogModuleTBPConnector, logger.LogLevelDebug, "recv msg dir %v, id %v", dir, id)

	return protocol.MSG_ID(xID), msgByteData, nil
}
