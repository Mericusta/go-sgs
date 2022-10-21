package protocol

import "fmt"

type Protocol any
type ProtocolID uint32

var protocolMakerMap map[ProtocolID]func() any = make(map[ProtocolID]func() any)

func RegisterProtocolMaker(id ProtocolID, f func() any) {
	protocolMakerMap[id] = f
}

func newProtocol(id ProtocolID) (any, error) {
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
