package connector

import (
	"net"

	"github.com/Mericusta/go-sgs/protocol"
)

type Connector interface {
	SendMsg(protocol.ProtocolID, protocol.Protocol) error
	RecvMsg() (protocol.ProtocolID, protocol.Protocol, error)
	Close() error
}

type BaseConnector struct {
	Connection net.Conn
}

func (c *BaseConnector) Address() string {
	return c.Connection.RemoteAddr().String()
}

func (c *BaseConnector) Close() error {
	return c.Connection.Close()
}

func New(connection net.Conn) Connector {
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
