package dispatcher

import "github.com/Mericusta/go-sgs/linker"

type IContext interface {
	Linker() *linker.Linker
	Dispatcher() *Dispatcher
}
