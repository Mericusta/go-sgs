package middleware

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
)

type IContext interface {
	Link() *link.Link
}

type HandlerMiddleware interface {
	Do(IContext, *event.Event) bool
}
