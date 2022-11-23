package dispatcher

import (
	"github.com/Mericusta/go-sgs/event"
)

type HandlerMiddleware interface {
	Do(IContext, *event.Event) bool
}

type RecoverMiddleware interface {
	Recover() bool
}
