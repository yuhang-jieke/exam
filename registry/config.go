package registry

import (
	"time"
)

// ConsulConfig Consul配置
type ConsulConfig struct {
	Address         string            `mapstructure:"address" json:"address" yaml:"address"`                            // Consul服务器地址
	Token           string            `mapstructure:"token" json:"token" yaml:"token"`                                  // ACL token（可选）
	Scheme          string            `mapstructure:"scheme" json:"scheme" yaml:"scheme"`                               // HTTP scheme (http或https)
	ServiceName     string            `mapstructure:"service_name" json:"service_name" yaml:"service_name"`             // 服务名称
	ServiceID       string            `mapstructure:"service_id" json:"service_id" yaml:"service_id"`                   // 服务ID（为空则自动生成）
	ServicePort     int               `mapstructure:"service_port" json:"service_port" yaml:"service_port"`             // 服务端口
	ServiceAddress  string            `mapstructure:"service_addr" json:"service_addr" yaml:"service_addr"`             // 服务地址（为空则自动检测）
	TTL             string            `mapstructure:"ttl" json:"ttl" yaml:"ttl"`                                        // TTL持续时间
	CheckTimeout    string            `mapstructure:"check_timeout" json:"check_timeout" yaml:"check_timeout"`          // 健康检查超时
	DeregisterAfter string            `mapstructure:"deregister_after" json:"deregister_after" yaml:"deregister_after"` // 自动注销时间
	Tags            []string          `mapstructure:"tags" json:"tags" yaml:"tags"`                                     // 服务标签
	Meta            map[string]string `mapstructure:"meta" json:"meta" yaml:"meta"`                                     // 服务元数据
}

// DefaultConfig 返回默认配置
func DefaultConfig() *ConsulConfig {
	return &ConsulConfig{
		Address:         "localhost:8500",
		Scheme:          "http",
		TTL:             "5s",
		CheckTimeout:    "3s",
		DeregisterAfter: "30s",
		Tags:            []string{},
		Meta:            make(map[string]string),
	}
}

// GetTTLDuration 解析TTL持续时间
func (c *ConsulConfig) GetTTLDuration() time.Duration {
	d, err := time.ParseDuration(c.TTL)
	if err != nil {
		return 5 * time.Second
	}
	return d
}

// GetCheckTimeoutDuration 解析健康检查超时
func (c *ConsulConfig) GetCheckTimeoutDuration() time.Duration {
	d, err := time.ParseDuration(c.CheckTimeout)
	if err != nil {
		return 3 * time.Second
	}
	return d
}

// GetDeregisterAfterDuration 解析自动注销时间
func (c *ConsulConfig) GetDeregisterAfterDuration() time.Duration {
	d, err := time.ParseDuration(c.DeregisterAfter)
	if err != nil {
		return 30 * time.Second
	}
	return d
}
