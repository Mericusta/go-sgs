package main

import (
	"github.com/Mericusta/go-sgs/dispatcher"
	"github.com/Mericusta/go-sgs/link"
)

type Client struct {
	*link.Link
	d    *dispatcher.Dispatcher
	data *userData
}

type userData struct {
	index     int
	expectMap map[int]int
}

func NewClient(l *link.Link, index int) *Client {
	return &Client{
		Link: l,
		// d:    dispatcher.New(), // NOTE: use global variable
		data: &userData{
			index:     index,
			expectMap: make(map[int]int),
		},
	}
}
