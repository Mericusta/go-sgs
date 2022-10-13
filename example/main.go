package main

import (
	"encoding/binary"
	"io"
	"net"
)

type Server struct {
	listener net.Listener
	// robotManager *robot.InnerServerRobotManager
}

const DefaultServerAddress string = "127.0.0.1:6666"

func NewInnserServer() *Server {
	listener, listenError := net.Listen("tcp", DefaultServerAddress)
	if listener == nil || listenError != nil {
		return nil
	}

	return &Server{
		listener: listener,
		// robotManager: &robot.InnerServerRobotManager{
		// 	BaseRobotManager: &robot.BaseRobotManager{
		// 		RobotMap: make(map[int64]robotInterface.Robot),
		// 	},
		// },
	}
}

func (s *Server) Run() {
	for {
		connection, acceptError := s.listener.Accept()
		if acceptError != nil {
			if acceptError.(*net.OpError).Err == net.ErrClosed {
				return
			}
			continue
		}

		go handleRead(connection)
		go handleWrite(connection)

		// serverRobot := s.robotManager.NewServerRobot(connection)
		// if serverRobot == nil {
		// 	continue
		// }
		// go serverRobot.Run(connection.RemoteAddr().String())
	}
}

func (s *Server) Exit(exitOvertimeSeconds int) {
	s.listener.Close()
}

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

func handleRead(connection net.Conn) {
	for {
		tag, msgByteData, err := receiveMsg(connection)
		if err != nil {
			if err != io.EOF && err.(*net.OpError).Err != net.ErrClosed {

			}
			close(r.ReceiveChan)
			r.ReceiveChan = nil
			return
		}
		if tag != protocol.MSG_ID_S2C_HEART_BEAT {

		}
	}
}

func receiveMsg(connection net.Conn) (MSG_ID, []byte, error) {
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
	return protocol.MSG_ID(tag), msgByteData, nil
}

func handleWrite(connection net.Conn) {
	for {
		sendMsg, ok := <-r.SendChan
		if !ok {

			return
		}
		if sendMsg.MsgID != protocol.MSG_ID_C2S_HEART_BEAT {

		}
		err := r.Connector.SendMsg(sendMsg.MsgID, sendMsg.ByteData)
		if err != nil {
			if err.(*net.OpError).Err != net.ErrClosed {

			}
			return
		}
	}
}

func sendMsg(msgID protocol.MSG_ID, msgByteData []byte) error {
	msgByteDataLength := uint32(len(msgByteData))
	tlvPackMsg := make([]byte, TLVPacketDataTagSize+TLVPacketDataLengthSize+msgByteDataLength)

	// tlvPackMsg[0,TLVPacketDataTagSize]
	binary.BigEndian.PutUint32(tlvPackMsg, uint32(msgID))

	// tlvPackMsg[TLVPacketDataTagSize,TLVPacketDataTagSize+TLVPacketDataLengthSize]
	binary.BigEndian.PutUint32(tlvPackMsg[TLVPacketDataTagSize:], msgByteDataLength)

	// tlvPackMsg[TLVPacketDataTagSize+TLVPacketDataLengthSize:]
	copy(tlvPackMsg[TLVPacketDataTagSize+TLVPacketDataLengthSize:], msgByteData)

	_, writeError := c.Connection.Write(tlvPackMsg)
	if writeError != nil {
		if writeError != io.EOF {

		}
		return writeError
	}

	return nil
}
