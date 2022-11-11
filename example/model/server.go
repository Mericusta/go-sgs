package model

type User struct {
	coutner int
}

func NewUser() *User {
	return &User{}
}

func (u *User) CounterIncrease() {
	u.coutner++
}

func (u *User) Counter() int {
	return u.coutner
}
