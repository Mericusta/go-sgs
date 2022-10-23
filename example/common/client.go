package common

import (
	"github.com/Mericusta/go-sgs/link"
)

type Client struct {
	*link.Link
	data *clientUserData
}

type clientUserData struct {
	index       int
	expectCount int
	expectMap   map[int]int
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

func (c *Client) Index() int {
	return c.data.index
}

func (c *Client) AddExpect(v int) int {
	k := c.data.expectCount
	c.data.expectMap[c.data.expectCount] = v
	c.data.expectCount++
	return k
}

func (c *Client) GetExpect(k int) int {
	return c.data.expectMap[k]
}

func (c *Client) DelExpect(k int) {
	delete(c.data.expectMap, k)
}
