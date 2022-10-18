package msg

import "github.com/Mericusta/go-sgs/protocol"

// Msg: contain protocol unique ID and protocol instance to transport from recv/send goroutine to logic goroutine
type Msg struct {
	id   protocol.ProtocolID
	data protocol.Protocol
}

func New(id protocol.ProtocolID, data protocol.Protocol) *Msg {
	return &Msg{id: id, data: data}
}

func (m *Msg) ID() protocol.ProtocolID {
	return m.id
}

func (m *Msg) Data() protocol.Protocol {
	return m.data
}
