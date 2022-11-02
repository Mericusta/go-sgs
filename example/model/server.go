package model

import "math/rand"

type User struct {
	Counter int `json:"counter"`
}

func NewUser() *User {
	return &User{Counter: rand.Intn(1024)}
}

func (u *User) AddCounter() {
	u.Counter++
}

func (u *User) GetCounter() int {
	return u.Counter
}
