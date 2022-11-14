package packer

import (
	"net"

	"github.com/Mericusta/go-sgs/protocol"
)

type Packer interface {
	SendMsg(protocol.ProtocolID, protocol.Protocol) error
	RecvMsg() (protocol.ProtocolID, protocol.Protocol, error)
	Close() error
}

type BasePacker struct {
	Connection net.Conn
}

func (c *BasePacker) Address() string {
	return c.Connection.RemoteAddr().String()
}

func (c *BasePacker) Close() error {
	return c.Connection.Close()
}

func New(connection net.Conn) Packer {
	return &MessagePacker{
		BasePacker: BasePacker{
			Connection: connection,
		},
	}
}
