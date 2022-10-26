package server

type User struct {
	counter int
}

func NewUser() *User {
	return &User{}
}

func (u *User) AddCounter() {
	u.counter++
}
