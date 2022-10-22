package common

import (
	"github.com/Mericusta/go-sgs/link"
)

type Client struct {
	*link.Link
	data *clientUserData
}

type clientUserData struct {
	index     int
	expectMap map[int]int
}

func NewClient(l *link.Link, index int) *Client {
	return &Client{
		Link: l,
		data: &clientUserData{
			index:     index,
			expectMap: make(map[int]int),
		},
	}
}
