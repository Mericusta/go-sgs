package connector

import (
	"net"
)

// Connector
// logic goroutine -msg-> connector -> marshaler -bytes-> pack -packet-> OS
// OS -packet-> unpack -bytes-> unmarshaler -> connector -msg-> logic goroutine

type Connector interface {
}

// type

type BaseConnector struct {
	net.Conn
}

func (c *BaseConnector) Address() string {
	return c.Conn.RemoteAddr().String()
}

func (c *BaseConnector) Close() {
	c.Conn.Close()
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
