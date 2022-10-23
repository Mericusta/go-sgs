package common

import (
	"github.com/Mericusta/go-sgs/link"
)

type User struct {
	*link.Link
	data *serverUserData
}

type serverUserData struct {
	index      int
	addCounter int
}

func NewUser(l *link.Link, index int) *User {
	return &User{
		Link: l,
		data: &serverUserData{
			index: index,
		},
	}
}

func (c *User) Index() int {
	return c.data.index
}

func (c *User) AddCounter() {
	c.data.addCounter++
}
