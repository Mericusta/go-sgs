package config

import "time"

const (
	DefaultServerAddress  string        = "127.0.0.1:6666"
	MaxConnectionCount    int           = 4096
	DispatcherLinkerCount int           = 128
	DispatcherCount       int           = MaxConnectionCount / DispatcherLinkerCount
	ChannelBuffer         int           = 8
	TcpKeepAlive          time.Duration = time.Second * 5
	TcpDialOvertime       time.Duration = time.Second * 30
)
