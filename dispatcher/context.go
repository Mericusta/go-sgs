package dispatcher

import "github.com/Mericusta/go-sgs/event"

type IContext interface {
	// Linker(uint64) *linker.Linker
	// Dispatcher() *Dispatcher
	UID() uint64
	Send(*event.Event)
	Exit()
}
