package dispatcher

import (
	"github.com/Mericusta/go-sgs/event"
)

type HandleMiddleware interface {
	Do(IContext, *event.Event) bool
}
