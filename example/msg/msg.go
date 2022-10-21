package msg

import "github.com/Mericusta/go-sgs/protocol"

const (
	C2SMsgID_CalculatorAdd = iota + 1
	S2CMsgID_CalculatorAdd
	C2SMsgID_CalculatorSub
	S2CMsgID_CalculatorSub
	C2SMsgID_CalculatorMul
	S2CMsgID_CalculatorMul
	C2SMsgID_CalculatorDiv
	S2CMsgID_CalculatorDiv
)

type C2SCalculatorData struct {
	Key    int `json:"key"`
	Value1 int `json:"value1"`
	Value2 int `json:"value2"`
}

type S2CCalculatorData struct {
	Key    int `json:"key"`
	Result int `json:"result"`
}

func init() {
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_CalculatorAdd), func() any { return &C2SCalculatorData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_CalculatorSub), func() any { return &C2SCalculatorData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_CalculatorMul), func() any { return &C2SCalculatorData{} })
	protocol.RegisterProtocolMaker(protocol.ProtocolID(C2SMsgID_CalculatorDiv), func() any { return &C2SCalculatorData{} })
}
