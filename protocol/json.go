// go:build tlv || json

package protocol

import (
	"encoding/json"
)

func Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func Unmarshal(id ProtocolID, b []byte) (any, error) {
	msg, err := newMsg(id)
	if msg == nil || err != nil {
		return nil, err
	}
	err = json.Unmarshal(b, msg)
	return msg, err
}
