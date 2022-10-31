package middleware

import (
	"net"

	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
)

type IContext interface {
	Link() *link.Link
}

type HandleMiddleware interface {
	Do(IContext, *event.Event) bool
}

type AcceptMiddleware interface {
	Do() (net.Conn, error, bool)
}
