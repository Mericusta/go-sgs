package model

import "time"

type Robot struct {
	id      uint64
	expect  map[int]int
	counter int
	t1      time.Time
	t2      time.Time
}

func NewRobot(id uint64) *Robot {
	return &Robot{id: id, expect: make(map[int]int)}
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

func (r *Robot) TickBegin() {
	r.t1 = time.Now()
}

func (r *Robot) TickEnd() {
	r.t2 = time.Now()
}
