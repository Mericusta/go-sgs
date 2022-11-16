package packer

import (
	"net"

	"github.com/Mericusta/go-sgs/protocol"
)

type Packer interface {
	// unpack into []byte from tcp packets and unmarshal []byte to memory data
	Unpack() (protocol.ProtocolID, protocol.ProtocolMsg, error)

	// marshal memory data to []byte and pack []byte into tcp packets
	Pack(protocol.ProtocolID, protocol.ProtocolMsg) error

	// tcp socket remote address
	Address() string

	// close tcp socket
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
