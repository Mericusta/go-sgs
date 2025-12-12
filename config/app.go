package config

import "fmt"

// 服务器基于项目的配置
type ServerAppConfig struct {
	Name          string               `mapstructure:"name"`          // 名称
	Protocol      *ProtocolConfig      `mapstructure:"protocol"`      // 协议
	Authorization *AuthorizationConfig `mapstructure:"authorization"` // 秘钥
}

func (c *ServerAppConfig) Check() error {
	if c == nil {
		return fmt.Errorf("server app config not exits")
	}
	if len(c.Name) == 0 {
		return fmt.Errorf("server app name is empty")
	}
	var err error
	err = c.Protocol.Check()
	if err != nil {
		return err
	}
	err = c.Authorization.Check()
	if err != nil {
		return err
	}
	return nil
}

type ProtocolConfig struct {
	MapPath string `mapstructure:"map_path"`
}

func (c *ProtocolConfig) Check() error {
	return nil
}

type AuthorizationConfig struct {
	Event string `mapstructure:"event"`
}

func (c *AuthorizationConfig) Check() error {
	return nil
}
