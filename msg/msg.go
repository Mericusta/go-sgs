package msg

type MsgID uint32

type Msg interface {
	Marshal() ([]byte, error)
}

func Unmarshal(msgID MsgID, bytes []byte) (Msg, error) {
	return nil, nil
}
