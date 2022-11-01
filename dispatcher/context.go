package dispatcher

import "github.com/Mericusta/go-sgs/link"

type IContext interface {
	Link() *link.Link
	Dispatcher() *Dispatcher
}
