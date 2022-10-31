package msg

import (
	serverModel "github.com/Mericusta/go-sgs/example/model/server"
)

const (
	C2SMsgID_Login = iota + 1
	S2CMsgID_Login
	C2SMsgID_Business
	S2CMsgID_Business
	C2SMsgID_Logout
	S2CMsgID_Logout
)

type C2SLoginData struct {
	AccountID int `json:"account_id"`
}

type S2CLoginData struct {
	User *serverModel.User `json:"user"`
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
