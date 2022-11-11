package msg

import (
	"github.com/Mericusta/go-sgs/protocol"
)

const (
	C2SMsgID_Login = iota + 1
	S2CMsgID_Login
	C2SMsgID_Business
	S2CMsgID_Business
	C2SMsgID_Logout
	S2CMsgID_Logout
)

func Init() {
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_Login), func() any { return &C2SLoginData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(S2CMsgID_Login), func() any { return &S2CLoginData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_Business), func() any { return &C2SBusinessData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(S2CMsgID_Business), func() any { return &S2CBusinessData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_Logout), func() any { return &C2SLogout{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(S2CMsgID_Logout), func() any { return &S2CLogout{} })
}

type C2SLoginData struct {
	AccountID uint64 `json:"account_id"`
}

type S2CLoginData struct {
	User *User `json:"user"`
}

type User struct {
	Counter int `json:"counter"`
}

type C2SBusinessData struct {
	Key    int `json:"key"`
	Value1 int `json:"value1"`
	Value2 int `json:"value2"`
}

type S2CBusinessData struct {
	Key    int `json:"key"`
	Result int `json:"result"`
}

type C2SLogout struct{}

type S2CLogout struct{}
