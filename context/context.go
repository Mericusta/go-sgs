package context

import "github.com/Mericusta/go-sgs/link"

type DispatchContext interface {
	Link() *link.Link
}

type BaseDispatchContext struct {
	l *link.Link
}

func (ctx *BaseDispatchContext) Link() *link.Link {
	return ctx.l
}

func NewDispatchContext(l *link.Link) DispatchContext {
	return &BaseDispatchContext{l: l}
}
