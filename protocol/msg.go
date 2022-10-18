package protocol

type ProtocolID uint32

type Protocol interface {
	Marshal() ([]byte, error)
}

func Unmarshal(protocolID ProtocolID, bytes []byte) (Protocol, error) {
	return nil, nil
}
