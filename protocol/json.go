//go:build json

package protocol

import (
	"encoding/json"
)

func Marshal(v ProtocolMsg) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(id ProtocolID, b []byte) (ProtocolMsg, error) {
	msg, err := newProtocolMsg(id)
	if msg == nil || err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, msg)
	return msg, err
}
