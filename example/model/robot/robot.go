package robot

import (
	"github.com/Mericusta/go-sgs/dispatcher"
)

type Robot struct {
	*dispatcher.Dispatcher
	id      uint64
	expect  map[int]int
	counter int
}

func NewRobot(id uint64) *Robot {
	return &Robot{id: id}
}

func (r *Robot) ID() uint64 {
	return r.id
}

func (r *Robot) AddExpect(k, v int) {
	r.expect[k] = v
}

func (r *Robot) GetExpect(k int) (int, bool) {
	v, has := r.expect[k]
	return v, has
}

func (r *Robot) DelExpect(k int) {
	delete(r.expect, k)
}

func (r *Robot) SetCounter(c int) {
	r.counter = c
}
