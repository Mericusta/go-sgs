package middleware

import (
	"github.com/Mericusta/go-sgs/event"
	"github.com/Mericusta/go-sgs/link"
)

type HandlerMiddleware []func(*link.Link, *event.Event)
