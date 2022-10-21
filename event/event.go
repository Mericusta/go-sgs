package event

import "github.com/Mericusta/go-sgs/protocol"

// Event: contain protocol unique ID and protocol instance to transport from recv/send goroutine to logic goroutine
type Event struct {
	id   protocol.ProtocolID
	data protocol.Protocol
}

func New(id protocol.ProtocolID, data protocol.Protocol) *Event {
	return &Event{id: id, data: data}
}

func (m *Event) ID() protocol.ProtocolID {
	return m.id
}

func (m *Event) Data() protocol.Protocol {
	return m.data
}
