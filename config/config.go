package config

import "time"

const (
	DefaultServerAddress string        = "192.168.2.203:6666"
	MaxConnectionCount   int           = 4096
	ChannelBuffer        int           = 8
	TcpKeepAlive         time.Duration = time.Second * 5
	TcpDialOvertime      time.Duration = time.Second * 30
)
