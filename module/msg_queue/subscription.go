package moduleMsgQueue

import (
	"github.com/Mericusta/go-sgs"
	"github.com/nats-io/nats.go"
)

// 使用接口的方式防止具体逻辑中直接引用该包

type ISubscribeBehavior interface {
	Subject() string
	Handler() func(sgs.Module, *nats.Msg)
}

type IUnsubscribeBehavior interface {
	Subject() string
}
