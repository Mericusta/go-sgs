package middleware

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
)

type HandlerMiddleware interface {
	Do(*link.Link, *event.Event) bool
}
