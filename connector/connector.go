package connector

import (
	"net"
	"robot-prototype/global"
	robotInterface "robot-prototype/interface"
	"time"
)

type BaseConnector struct {
	Connection net.Conn
}

func (c *BaseConnector) Address() string {
	return c.Connection.RemoteAddr().String()
}

func (c *BaseConnector) Close() {
	c.Connection.Close()
}

func NewConnector(connection net.Conn) robotInterface.Connector {
	return &MessageConnector{
		BaseConnector: BaseConnector{
			Connection: connection,
		},
	}
}

func NewConnectorWithAddress(address string) robotInterface.Connector {
	connection, dailError := net.DialTimeout("tcp", address, time.Second*time.Duration(global.ConnectorDialTimeoutSeconds))
	if dailError != nil {
		return nil
	}
	return &MessageConnector{
		BaseConnector: BaseConnector{
			Connection: connection,
		},
	}
}
