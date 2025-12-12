package config

import "fmt"

type MsgQueueConfig struct {
	URL    string            `mapstructure:"url"`
	Option map[string]string `mapstructure:"option"`
}

func (c *MsgQueueConfig) Check() error {
	if c == nil {
		return fmt.Errorf("MsgQueueConfig is nil")
	}
	return nil
}
