package server

import "math/rand"

type User struct {
	counter int
}

func NewUser() *User {
	return &User{counter: rand.Intn(1024)}
}

func (u *User) AddCounter() {
	u.counter++
}

func (u *User) GetCounter() int {
	return u.counter
}
