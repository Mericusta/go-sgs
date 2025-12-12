package sgs

import (
	"os"
	"strconv"
)

// 服务器基础配置
type ServerBasicConfig struct {
	Debug            bool                    `mapstructure:"debug"`             // 开启 debug 模式
	LogLevel         string                  `mapstructure:"log_level"`         // 日志等级
	PPROFPort        int32                   `mapstructure:"pprof_port"`        // PPROF 端口
	ServiceDiscovery *ServiceDiscoveryConfig `mapstructure:"service_discovery"` // 服务发现配置
	Monitor          *MonitorConfig          `mapstructure:"monitor"`           // 监控
}

type ServiceDiscoveryConfig struct {
	ServerIDKey      string `mapstructure:"server_id_key"`      // 服务发现分配 ID 的 key
	ReportKey        string `mapstructure:"report_key"`         // 服务发现上报的 key
	PublishKey       string `mapstructure:"publish_key"`        // 服务发现状态变更广播的 key
	KeepAliveSeconds int64  `mapstructure:"keep_alive_seconds"` // 服务发现保活间隔
	ApiVersion       string `mapstructure:"api_version"`        // 服务器版本
}

func (c *ServerBasicConfig) Check() error {
	if c == nil {
		c = &ServerBasicConfig{}
	}
	if len(c.LogLevel) == 0 {
		c.LogLevel = "info"
	}
	// 有环境变量优先使用环境变量
	pprofPortStr := os.Getenv("PPROF_PORT")
	if len(pprofPortStr) > 0 {
		pprofPort, err := strconv.ParseInt(pprofPortStr, 10, 64)
		if err != nil {
			return err
		}
		c.PPROFPort = int32(pprofPort)
	}

	apiVersion := os.Getenv("API_VERSION")
	if len(apiVersion) > 0 {
		c.ServiceDiscovery.ApiVersion = apiVersion
	}
	return nil
}

func (c *ServerBasicConfig) GetApiVersion() string {
	return c.ServiceDiscovery.ApiVersion
}

func (c *ServerBasicConfig) GetReportKey() string {
	if len(c.ServiceDiscovery.ApiVersion) == 0 {
		return c.ServiceDiscovery.ReportKey
	}
	return c.ServiceDiscovery.ReportKey + "_v" + c.ServiceDiscovery.ApiVersion
}

type MonitorConfig struct {
	UnmountReportMark      string `mapstructure:"unmount_report_mark"`      // 主服务卸载通知的标识
	UnmountNotificationUrl string `mapstructure:"unmount_notification_url"` // 主服务卸载时通知
}
