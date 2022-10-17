package connector

import (
	"net"

	"github.com/Mericusta/go-sgs/msg"
)

// Connector
// logic goroutine -msg-> connector -> marshaler -bytes-> pack -packet-> OS
// OS -packet-> unpack -bytes-> unmarshaler -> connector -msg-> logic goroutine

type Connector interface {
	SendMsg(msg.MsgID, msg.Msg) error
}

type BaseConnector struct {
	Connection net.Conn
}

func (c *BaseConnector) Address() string {
	return c.Connection.RemoteAddr().String()
}

func (c *BaseConnector) Close() {
	c.Connection.Close()
}

func NewConnector(connection net.Conn) Connector {
	return &MessageConnector{
		BaseConnector: BaseConnector{
			Connection: connection,
		},
	}
}

// func NewConnectorWithAddress(address string) robotInterface.Connector {
// 	connection, dailError := net.DialTimeout("tcp", address, time.Second*time.Duration(global.ConnectorDialTimeoutSeconds))
// 	if dailError != nil {
// 		return nil
// 	}
// 	return &MessageConnector{
// 		BaseConnector: BaseConnector{
// 			Connection: connection,
// 		},
// 	}
// }
