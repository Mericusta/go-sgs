package protocol

import "fmt"

type ProtocolMsg any
type ProtocolID uint32

var protocolMakerMap map[ProtocolID]func() ProtocolMsg = make(map[ProtocolID]func() ProtocolMsg)

func RegisterProtocolMaker(id ProtocolID, f func() ProtocolMsg) {
	protocolMakerMap[id] = f
}

func newProtocolMsg(id ProtocolID) (ProtocolMsg, error) {
	maker := protocolMakerMap[id]
	if maker == nil {
		return nil, fmt.Errorf("unknown protocol id %v", id)
	}
	p := maker()
	if p == nil {
		return nil, fmt.Errorf("protocol id %v maker make a nil protocol", id)
	}
	return p, nil
}
